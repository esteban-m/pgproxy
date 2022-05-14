package inbound_mysql

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EngineersBox/hexdump-format"
)

type MySQLConnState int32
type EndpointSide string

type SQLSnifferHandler func(sql string)
type UnknownPacketHandler func(packet *RawPacket, side EndpointSide)

const (
	StateConn    MySQLConnState = 1
	StateCommand MySQLConnState = 2
)

const (
	TypeServerSide EndpointSide = "server"
	TypeClientSide EndpointSide = "client"
)

type MySQLConn struct {
	State MySQLConnState

	CommandQueue list.List
	QueueLock    sync.Mutex

	SQLSnifferHandler    SQLSnifferHandler
	UnknownPacketHandler UnknownPacketHandler

	DebugUnknownPacketType bool
	DebugRawPacket         bool
}

func NewMySQLConn() *MySQLConn {
	return &MySQLConn{
		State:                StateConn,
		SQLSnifferHandler:    func(sql string) {},
		UnknownPacketHandler: func(packet *RawPacket, side EndpointSide) {},

		DebugRawPacket: true,
	}
}

func NewMySQLConnWithSniffer(handler SQLSnifferHandler, unknownPacketHandler UnknownPacketHandler) *MySQLConn {
	c := NewMySQLConn()
	c.SQLSnifferHandler = handler
	c.UnknownPacketHandler = unknownPacketHandler

	return c
}

func (this *MySQLConn) Enqueue(rawPacket *RawPacket) {
	this.QueueLock.Lock()

	this.CommandQueue.PushBack(rawPacket)

	this.QueueLock.Unlock()
}

func (this *MySQLConn) DequeuePacket() *RawPacket {
	this.QueueLock.Lock()

	elem := this.CommandQueue.Front()
	if elem != nil {
		this.CommandQueue.Remove(elem)
	}

	this.QueueLock.Unlock()

	if elem != nil {
		return elem.Value.(*RawPacket)
	} else {
		return nil
	}
}

func (this *MySQLConn) ChangeState(newState MySQLConnState) (MySQLConnState, bool, error) {
	switch this.GetCurrentState() {
	case StateConn:
		switch newState {
		case StateCommand:
			ok := atomic.CompareAndSwapInt32((*int32)(&this.State), int32(StateConn), int32(newState))
			return newState, ok, nil
		default:
			return 0, false, errors.New(fmt.Sprint("conn state -> ", newState, " cannot be performed"))
		}
	case StateCommand:
		return this.GetCurrentState(), false, nil
	default:
		return 0, false, errors.New(fmt.Sprint("unknown state:", this.State))
	}
}

func (this *MySQLConn) GetCurrentState() MySQLConnState {
	return MySQLConnState(atomic.LoadInt32((*int32)(&this.State)))
}

func (this *MySQLConn) HandleCommandPhaseServerPacket(rawPacket *RawPacket) error {
	req := this.DequeuePacket()

	return this.handleServerPacketInternal(rawPacket, req)
}

func (this *MySQLConn) handleServerPacketInternal(rawPacket *RawPacket, requestRawPacket *RawPacket) error {
	p, err := ParseServerPayloadPacket(rawPacket.Data, rawPacket.SeqId)
	if err != nil {
		return err
	}

	switch p.PacketType() {
	case CommonOkPacketType:
		return nil
	case UnknownPacketType:
		np := p.(*UnknownPacket)
		if this.DebugUnknownPacketType {
			fmt.Println("unknown server packet type:", np.Type, "length:", len(np.Data))
		}
		this.UnknownPacketHandler(rawPacket, TypeServerSide)
		return nil
	default:
		return nil
	}
}

func (this *MySQLConn) HandleCommandPhaseClientPacket(rawPacket *RawPacket) error {

	err := this.handleClientPacketInternal(rawPacket)
	if err != nil {
		return err
	} else {
		this.Enqueue(rawPacket)
		return nil
	}
}

func (this *MySQLConn) handleClientPacketInternal(rawPacket *RawPacket) error {
	p, err := ParseClientPayloadPacket(rawPacket.Data)
	if err != nil {
		return err
	}

	switch p.PacketType() {
	case ClientCommandStatementPrepareType:
		np := p.(*ClientCommandStatementPreparePacket)
		fmt.Println("prepared statement:[" + np.Statement + "]")
		this.SQLSnifferHandler(np.Statement)
		return nil
	case ClientCommandQueryType:
		np := p.(*ClientCommandQueryPacket)
		fmt.Println("query:[" + np.Statement + "]")
		this.SQLSnifferHandler(np.Statement)
		return nil
	case UnknownPacketType:
		np := p.(*UnknownPacket)
		if this.DebugUnknownPacketType {
			fmt.Println("unknown client packet type:", np.Type, "length:", len(np.Data))
		}
		this.UnknownPacketHandler(rawPacket, TypeClientSide)
		return nil
	default:
		return nil
	}
}

func (this *MySQLConn) HandleConnPhase(inbound, outbound io.ReadWriter) error {
	// handle conn phase including auth methods

	// server init & methods
	//FIXME current auth flow:
	// 1. server -> client -> server ok (wordpress)
	// 2. server -> client -> server -> server ok
	{
		// server side
		err := this.connPhaseServerSide(inbound, outbound)
		if err != nil {
			return err
		}

		// client side
		err = this.connPhaseClientSide(inbound, outbound)
		if err != nil {
			return err
		}

		// server side
		err = this.connPhaseServerSide(inbound, outbound)
		if err != nil {
			return err
		}

		// server side
		if this.GetCurrentState() != StateCommand {
			err = this.connPhaseServerSide(inbound, outbound)
			if err != nil {
				return err
			}
		}

		if this.GetCurrentState() != StateCommand {
			return errors.New("connection state is not successfully changed")
		}
	}

	return nil
}
func (this *MySQLConn) HandleConnPhase0(inbound, outbound io.ReadWriter) error {
	// handle conn phase including auth methods

	// server init & methods
	//FIXME current auth flow:
	// 1. server -> client -> server ok (wordpress)
	// 2. server -> client -> server -> server ok
	go func() {
		for {
			err := this.connPhaseServerSide(inbound, outbound)
			if err != nil {
				panic(err)
			}
		}
	}()
	go func() {
		for {
			err := this.connPhaseClientSide(inbound, outbound)
			if err != nil {
				panic(err)
			}
		}
	}()

	time.Sleep(1 * time.Hour)

	return nil
}

func (this *MySQLConn) connPhaseClientSide(inbound io.ReadWriter, outbound io.ReadWriter) error {
	{
		rawPacket := new(RawPacket)
		err := rawPacket.ReadRawPacket(inbound)
		if err != nil {
			fmt.Println("source -> dest: read error:", err)
			return errors.New("read raw packet error from inbound read")
		}

		if this.DebugRawPacket {
			fmt.Println("recv connPhaseClientSide packet length:", len(rawPacket.Data))
			fmt.Println("recv connPhaseClientSide packet :", hexdump.CreateHexdumpText(rawPacket.Data))
		}

		err = rawPacket.WriteRawPacket(outbound)
		if err != nil {
			fmt.Println("source -> dest: write error:", err)
			return errors.New("write raw packet error to outbound read")
		}
	}
	return nil
}

func (this *MySQLConn) connPhaseServerSide(inbound io.ReadWriter, outbound io.ReadWriter) error {
	rawPacket := new(RawPacket)
	err := rawPacket.ReadRawPacket(outbound)
	if err != nil {
		fmt.Println("dest -> source: read error:", err)
		return errors.New("read raw packet error from outbound read")
	}

	if this.DebugRawPacket {
		fmt.Println("recv connPhaseServerSide packet length:", len(rawPacket.Data))
		fmt.Println("recv connPhaseServerSide packet :", hexdump.CreateHexdumpText(rawPacket.Data))
	}

	p, err := ParseServerPayloadPacket(rawPacket.Data, rawPacket.SeqId)
	if err != nil {
		return err
	}
	if p.PacketType() == CommonOkPacketType {
		// change conn state
		if this.GetCurrentState() == StateConn {
			if _, _, err := this.ChangeState(StateCommand); err != nil {
				return err
			}
			fmt.Println("change conn state to command phase")
		}
	}

	err = rawPacket.WriteRawPacket(inbound)
	if err != nil {
		fmt.Println("dest -> source: write error:", err)
		return errors.New("write raw packet error to inbound read")
	}
	return nil
}

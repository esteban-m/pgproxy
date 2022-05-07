package main

import (
	"errors"
	"fmt"
	"sync/atomic"
)

type MySQLConnState int32

const (
	StateConn    MySQLConnState = 1
	StateCommand MySQLConnState = 2
)

type MySQLConn struct {
	State MySQLConnState

	DebugUnknownPacketType bool
}

func NewMysSQLConn() *MySQLConn {
	return &MySQLConn{State: StateConn}
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

func (this *MySQLConn) HandleServerPacket(packet *RawPacket) error {
	p, err := ParseServerPayloadPacket(packet.Data, packet.SeqId)
	if err != nil {
		return err
	}

	switch p.PacketType() {
	case CommonOkPacketType:
		// change conn state
		if this.GetCurrentState() == StateConn {
			if _, _, err := this.ChangeState(StateCommand); err != nil {
				return err
			}
		}
		return nil
	case UnknownPacketType:
		np := p.(*UnknownPacket)
		if this.DebugUnknownPacketType {
			fmt.Println("unknown server packet type:", np.Type, "length:", len(np.Data))
		}
		//TODO
		return nil
	default:
		return nil
	}
}

func (this *MySQLConn) HandleClientPacket(packet *RawPacket) error {
	p, err := ParseClientPayloadPacket(packet.Data)
	if err != nil {
		return err
	}

	switch p.PacketType() {
	case ClientCommandStatementPrepareType:
		np := p.(*ClientCommandStatementPreparePacket)
		fmt.Println("prepared statement:[" + np.Statement + "]")
		return nil
	case ClientCommandQueryType:
		np := p.(*ClientCommandQueryPacket)
		fmt.Println("query:[" + np.Statement + "]")
		return nil
	case UnknownPacketType:
		np := p.(*UnknownPacket)
		if this.DebugUnknownPacketType {
			fmt.Println("unknown client packet type:", np.Type, "length:", len(np.Data))
		}
		//TODO
		return nil
	default:
		return nil
	}
}

package inbound_mysql

import (
	"encoding/binary"
	"errors"
)

var ErrPacketIncomplete = errors.New("packet is incomplete")
var ErrReaderEOF = errors.New("reader EOF")
var ErrUnknownPacket = errors.New("unknown packet")

type PacketType int

const (
	UnknownPacketType  PacketType = -1
	CommonOkPacketType PacketType = 1
)

const (
	ClientCommandQueryType            PacketType = 101
	ClientCommandStatementPrepareType PacketType = 102
	ClientCommandStatementCloseType   PacketType = 103
)

const (
	ServerCommandStatementResponseType PacketType = 201
)

type Packet interface {
	PacketType() PacketType
}

func ParseServerPayloadPacket(data []byte, seqId uint8) (Packet, error) {
	switch data[0] {
	case 0x00:
		fallthrough
	case 0xFE:
		return new(OkPacket), nil
	default:
		return &UnknownPacket{
			Type: data[0],
			Data: data,
		}, nil
	}
}

func ParseClientPayloadPacket(data []byte) (Packet, error) {
	switch data[0] {
	case 0x03:
		p := new(ClientCommandQueryPacket)
		p.Statement = string(data[1:])
		return p, nil
	case 0x16:
		p := new(ClientCommandStatementPreparePacket)
		p.Statement = string(data[1:])
		return p, nil
	case 0x19:
		p := new(ClientCommandStatementClosePacket)
		p.StatementId = binary.LittleEndian.Uint32(data[1:])
		return p, nil
	default:
		return &UnknownPacket{
			Type: data[0],
			Data: data,
		}, nil
	}
}

type UnknownPacket struct {
	Type uint8
	Data []byte
}

func (this *UnknownPacket) PacketType() PacketType {
	return UnknownPacketType
}

type OkPacket struct {
	//TODO
}

func (o *OkPacket) PacketType() PacketType {
	return CommonOkPacketType
}

type ServerCommandStatementResponsePacket struct {
	StatementId uint32
	//TODO
}

func (o *ServerCommandStatementResponsePacket) PacketType() PacketType {
	return ServerCommandStatementResponseType
}

type ClientCommandQueryPacket struct {
	Statement string
}

func (this *ClientCommandQueryPacket) PacketType() PacketType {
	return ClientCommandQueryType
}

type ClientCommandStatementPreparePacket struct {
	Statement string
}

func (this *ClientCommandStatementPreparePacket) PacketType() PacketType {
	return ClientCommandStatementPrepareType
}

type ClientCommandStatementClosePacket struct {
	StatementId uint32
}

func (this *ClientCommandStatementClosePacket) PacketType() PacketType {
	return ClientCommandStatementCloseType
}

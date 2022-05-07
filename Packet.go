package main

import (
	"encoding/binary"
	"fmt"
	"github.com/EngineersBox/hexdump-format"
	"io"
	"strings"
)

type RawPacket struct {
	Length uint32
	SeqId  uint8
	Data   []byte
}

func (this *RawPacket) ReadRawPacket(reader io.Reader) error {
	arr := [4]byte{0, 0, 0, 0}
	_, err := io.ReadFull(reader, arr[:3])
	if err != nil {
		return err
	}

	this.Length = binary.LittleEndian.Uint32(arr[:])

	_, err = io.ReadFull(reader, arr[:1])
	if err != nil {
		return err
	}
	this.SeqId = uint8(arr[0])

	data := make([]byte, this.Length)

	_, err = io.ReadFull(reader, data)
	if err != nil {
		return err
	}
	this.Data = data

	return nil
}

func (this *RawPacket) WriteRawPacket(writer io.Writer) error {
	arr := [4]byte{0, 0, 0, 0}

	binary.LittleEndian.PutUint32(arr[:], this.Length)
	_, err := writer.Write(arr[:3])
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte{this.SeqId})
	if err != nil {
		return err
	}
	_, err = writer.Write(this.Data)
	if err != nil {
		return err
	}
	return nil
}

func (this *RawPacket) String() string {
	builder := new(strings.Builder)
	builder.WriteString(fmt.Sprintln("packet meta: size:", this.Length, ",", "seq:", this.SeqId))
	builder.WriteString(fmt.Sprintln(hexdump.CreateHexdumpText(this.Data)))
	return builder.String()
}

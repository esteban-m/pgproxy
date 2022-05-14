package inbound_mysql

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/EngineersBox/hexdump-format"
)

type RawPacket struct {
	Length uint32
	SeqId  uint8
	Data   []byte
}

func (this *RawPacket) ReadRawPacket(reader io.Reader) error {
	// TODO support packet size over 16MB
	arr := [4]byte{0, 0, 0, 0}
	_, err := io.ReadFull(reader, arr[:])
	if err != nil {
		return err
	}

	this.SeqId = arr[3]
	arr[3] = 0
	this.Length = binary.LittleEndian.Uint32(arr[:])

	data := make([]byte, this.Length)

	_, err = io.ReadFull(reader, data)
	if err != nil {
		return err
	}
	this.Data = data

	return nil
}

func (this *RawPacket) WriteRawPacket(writer io.Writer) error {
	// TODO support packet size over 16MB
	arr := [4]byte{0, 0, 0, 0}

	binary.LittleEndian.PutUint32(arr[:], this.Length)
	arr[3] = this.SeqId
	_, err := writer.Write(arr[:])
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

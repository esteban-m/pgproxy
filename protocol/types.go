package protocol

import "errors"

var _ErrNoData = errors.New("no data")

type Type interface {
	FromBytes(buf []byte) (rest []byte, fin bool, err error)
	ToBytes() []byte
	RequiredSize() int
}

var _ Type = new(Int8)

type Int8 struct {
	fin   bool
	Value byte
}

func (i *Int8) FromBytes(buf []byte) (rest []byte, fin bool, err error) {
	if len(buf) <= 0 {
		return nil, i.fin, _ErrNoData
	}
	i.Value = buf[0]
	i.fin = true
	return buf[1:], true, nil
}

func (i *Int8) ToBytes() []byte {
	return []byte{i.Value}
}

func (i *Int8) RequiredSize() int {
	return 1
}

type Int16 struct {
}

type Int32 struct {
}

type ArrInt8 []Int8

type ArrInt16 []Int16

type ArrInt32 []Int32

type String struct {
}

type Bytes struct {
}

package protocol

import "errors"

var _ErrNoData = errors.New("no data")

type Type interface {
	FromBytes(buf []byte) (rest []byte, fin bool, err error)
	ToBytes() []byte
	RequiredSize() int
}

var _ Type = new(Int8)
var _ Type = new(Int16)
var _ Type = new(Int32)

type Int8 struct {
	fin   bool
	Value byte
}

func (i *Int8) FromBytes(buf []byte) (rest []byte, fin bool, err error) {
	if len(buf) <= 0 {
		return nil, i.fin, nil
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
	Values [2]byte
	filled byte
}

func (i *Int16) FromBytes(buf []byte) (rest []byte, fin bool, err error) {
	bCnt := i.RequiredSize()
	max := bCnt - int(i.filled)
	if max > len(buf) {
		max = len(buf)
	}
	for idx := 0; idx < max; idx++ {
		i.Values[int(i.filled)] = buf[idx]
	}
	if int(i.filled) >= bCnt {
		return buf[max:], true, nil
	} else {
		return buf[max:], false, nil
	}
}

func (i *Int16) ToBytes() []byte {
	return i.Values[:]
}

func (i *Int16) RequiredSize() int {
	return 2
}

type Int32 struct {
	Values [4]byte
	filled byte
}

func (i *Int32) FromBytes(buf []byte) (rest []byte, fin bool, err error) {
	bCnt := i.RequiredSize()
	max := bCnt - int(i.filled)
	if max > len(buf) {
		max = len(buf)
	}
	for idx := 0; idx < max; idx++ {
		i.Values[int(i.filled)] = buf[idx]
	}
	if int(i.filled) >= bCnt {
		return buf[max:], true, nil
	} else {
		return buf[max:], false, nil
	}
}

func (i *Int32) ToBytes() []byte {
	return i.Values[:]
}

func (i *Int32) RequiredSize() int {
	return 4
}

type ArrInt8 []Int8

type ArrInt16 []Int16

type ArrInt32 []Int32

type String struct {
}

type Bytes struct {
}

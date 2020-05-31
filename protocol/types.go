package protocol

type Type interface {
	FromBytes(buf []byte) (rest []byte, fin bool, err error)
	ToBytes() []byte
	RequiredSize() int
}

type Int8 struct {
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

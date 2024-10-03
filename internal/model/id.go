package model

import "encoding/binary"

type ID uint64

func (i ID) ToBytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func (i ID) ToUint64() uint64 {
	return uint64(i)
}

func IDFromBytes(data []byte) ID {
	return ID(binary.BigEndian.Uint64(data))
}

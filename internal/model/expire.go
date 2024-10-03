package model

import (
	"encoding/binary"
	"time"
)

type ExpireData struct {
	Id         uint64 `json:"id"`
	ExpireTime uint64 `json:"expire_time"`
}

func NewExpireData(id uint64, expireTime time.Time) *ExpireData {
	return &ExpireData{
		Id:         id,
		ExpireTime: uint64(expireTime.UnixNano()),
	}
}

func (e *ExpireData) ToBytes() []byte {
	buf := make([]byte, 16)

	binary.BigEndian.PutUint64(buf[:8], e.Id)
	binary.BigEndian.PutUint64(buf[8:], e.ExpireTime)

	return buf
}

func ExpireDataFromBytes(data []byte) *ExpireData {
	id := binary.BigEndian.Uint64(data[:8])
	expireTime := binary.BigEndian.Uint64(data[8:])

	return &ExpireData{
		Id:         id,
		ExpireTime: expireTime,
	}
}

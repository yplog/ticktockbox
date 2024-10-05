package model

import "encoding/json"

type Data struct {
	Content map[string]interface{} `json:"content"`
}

func NewData(content map[string]interface{}) *Data {
	return &Data{
		Content: content,
	}
}

func (d *Data) ToBytes() []byte {
	return nil
}

func DataFromBytes(data []byte) *Data {
	var d Data

	err := json.Unmarshal(data, &d)
	if err != nil {
		return nil
	}

	return &d
}

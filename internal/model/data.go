package model

type Data struct {
	Content map[string]interface{} `json:"content"`
}

func NewData(content map[string]interface{}) *Data {
	return &Data{
		Content: content,
	}
}

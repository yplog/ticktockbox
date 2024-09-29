package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ItemStatus string

const (
	StatusActive  ItemStatus = "Active"
	StatusArchive ItemStatus = "Archive"
)

var ErrInvalidStatus = errors.New("invalid status")

type Item struct {
	ID         string     `json:"id"`
	Status     ItemStatus `json:"status"`
	ExpireDate time.Time  `json:"expireDate"`
	Data       string     `json:"data"`
}

func NewItem(expireDate time.Time, data string) *Item {
	id := uuid.New().String()

	return &Item{
		ID:         id,
		Status:     StatusActive,
		ExpireDate: expireDate,
		Data:       data,
	}
}

func (i *Item) MarshalJSON() ([]byte, error) {
	type Alias Item
	return json.Marshal(&struct {
		*Alias
		ExpireDate string `json:"expireDate"`
		Data       string `json:"data"`
	}{
		Alias:      (*Alias)(i),
		ExpireDate: i.ExpireDate.Format(time.RFC3339),
		Data:       i.Data,
	})
}

func (i *Item) UnmarshalJSON(data []byte) error {
	type Alias Item
	aux := &struct {
		*Alias
		ExpireDate string `json:"expireDate"`
		Data       string `json:"data"`
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	i.ExpireDate, err = time.Parse(time.RFC3339, aux.ExpireDate)
	if err != nil {
		return err
	}
	i.Data = aux.Data

	return nil
}

func (i *Item) SetStatus(status ItemStatus) error {
	if status != StatusActive && status != StatusArchive {
		return ErrInvalidStatus
	}
	i.Status = status
	return nil
}

func (i *Item) Key() []byte {
	return []byte(i.ID)
}

func (i *Item) Value() ([]byte, error) {
	return json.Marshal(i)
}

func ItemFromValue(value []byte) (*Item, error) {
	var item Item
	err := json.Unmarshal(value, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

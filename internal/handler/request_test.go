package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDecodeRequestBody(t *testing.T) {
	validData := RequestData{
		Expire: time.Now().Add(1 * time.Hour),
		Data:   map[string]interface{}{"key": "value"},
	}
	validBody, _ := json.Marshal(validData)

	tests := []struct {
		name          string
		body          []byte
		expectedError bool
	}{
		{
			name:          "Valid request",
			body:          validBody,
			expectedError: false,
		},
		{
			name:          "Invalid JSON",
			body:          []byte(`{"expire": "invalid-date", "data": {"key": "value"}}`),
			expectedError: true,
		},
		{
			name:          "Expire in the past",
			body:          []byte(`{"expire": "2000-01-01T00:00:00Z", "data": {"key": "value"}}`),
			expectedError: true,
		},
		{
			name:          "Expire is zero",
			body:          []byte(`{"expire": "0001-01-01T00:00:00Z", "data": {"key": "value"}}`),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.body))
			req.Header.Set("Content-Type", "application/json")

			_, err := DecodeRequestBody(req)
			if (err != nil) != tt.expectedError {
				t.Errorf("DecodeRequestBody() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}

func TestValidateExpire(t *testing.T) {
	tests := []struct {
		name          string
		expire        time.Time
		expectedError bool
	}{
		{
			name:          "Valid expire time",
			expire:        time.Now().Add(1 * time.Hour),
			expectedError: false,
		},
		{
			name:          "Invalid format",
			expire:        time.Time{},
			expectedError: true,
		},
		{
			name:          "Expire is zero",
			expire:        time.Time{},
			expectedError: true,
		},
		{
			name:          "Expire in the past",
			expire:        time.Now().Add(-1 * time.Hour),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExpire(tt.expire)
			if (err != nil) != tt.expectedError {
				t.Errorf("validateExpire() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}

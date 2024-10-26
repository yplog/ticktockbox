package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	WriteJSONResponse(rr, http.StatusOK, data)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected, _ := json.Marshal(data)
	if rr.Body.String() != string(expected)+"\n" { // Satır sonunu kontrol edin
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expected))
	}
}

func TestWriteErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	message := "error occurred"

	WriteErrorResponse(rr, http.StatusBadRequest, message)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected, _ := json.Marshal(map[string]interface{}{"error": message, "status": http.StatusBadRequest})
	if rr.Body.String() != string(expected)+"\n" { // Satır sonunu kontrol edin
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expected))
	}
}

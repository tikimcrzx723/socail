package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSON(rw http.ResponseWriter, status int, data any) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	return json.NewEncoder(rw).Encode(data)
}

func readJSON(rw http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1 mb
	r.Body = http.MaxBytesReader(rw, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func writeJSONError(rw http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return writeJSON(rw, status, &envelope{Error: message})
}

func (app *application) jsonResponse(rw http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return writeJSON(rw, status, &envelope{Data: data})
}

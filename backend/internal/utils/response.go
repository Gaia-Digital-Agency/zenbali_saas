package utils

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success sends a success response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created sends a created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// Error sends an error response
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

// BadRequest sends a 400 error
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 error
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 error
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Error(w, http.StatusForbidden, message)
}

// NotFound sends a 404 error
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Not found"
	}
	Error(w, http.StatusNotFound, message)
}

// InternalError sends a 500 error
func InternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(w, http.StatusInternalServerError, message)
}

// Message sends a simple message response
func Message(w http.ResponseWriter, message string) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
	})
}

// ParseJSON parses JSON request body
func ParseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

package internal

import (
	"errors"
	"net/http"
)

var (
	ErrSelfMessage  = errors.New("cannot message yourself")
	ErrPeerOffline  = errors.New("peer offline")
	ErrInvalidUser  = errors.New("invalid user")
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

type HTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func ToHTTP(err error) (int, HTTPError) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: err.Error()}
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrSelfMessage), errors.Is(err, ErrInvalidUser):
		return http.StatusBadRequest, HTTPError{Status: http.StatusBadRequest, Code: "bad_request", Message: err.Error()}
	default:
		return http.StatusInternalServerError, HTTPError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal server error"}
	}
}

type WSError struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToWS(err error) WSError {
	switch {
	case errors.Is(err, ErrPeerOffline):
		return WSError{Type: "error", Code: "peer_offline", Message: err.Error()}
	case errors.Is(err, ErrSelfMessage):
		return WSError{Type: "error", Code: "self_message", Message: err.Error()}
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidUser):
		return WSError{Type: "error", Code: "bad_request", Message: err.Error()}
	default:
		return WSError{Type: "error", Code: "internal_error", Message: "internal server error"}
	}
}

package stringsvc

import (
	"context"
	"errors"
	"strings"
)

// StringService is a concrete implementation of StringService.
type StringService struct{}

// Uppercase capitalizes words.
func (StringService) Uppercase(_ context.Context, s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	return strings.ToUpper(s), nil
}

// Count returns the length of strings.
func (StringService) Count(_ context.Context, s string) int {
	return len(s)
}

// ErrEmpty is returned when an input string is empty.
var ErrEmpty = errors.New("empty string")

// UppercaseRequest is a request for stringsvc.
type UppercaseRequest struct {
	S string `json:"s"`
}

// UppercaseResponse is the response to uppercase requests.
type UppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

// CountRequest is a request for strinsvc.
type CountRequest struct {
	S string `json:"s"`
}

// CountResponse is the response to count requests.
type CountResponse struct {
	V int `json:"v"`
}

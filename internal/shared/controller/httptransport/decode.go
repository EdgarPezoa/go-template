package httptransport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type InvalidRequestError struct {
	err error
}

func (requestError InvalidRequestError) Error() string { return requestError.err.Error() }
func (requestError InvalidRequestError) Unwrap() error { return requestError.err }

func Decode(responseWriter http.ResponseWriter, request *http.Request, target any) error {
	contentType := strings.TrimSpace(strings.Split(request.Header.Get("Content-Type"), ";")[0])
	if !strings.EqualFold(contentType, "application/json") {
		return InvalidRequestError{err: errors.New("content-type must be application/json")}
	}
	request.Body = http.MaxBytesReader(responseWriter, request.Body, 1<<20)
	defer request.Body.Close()
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return InvalidRequestError{err: errors.New("request body must not be empty")}
		}
		return InvalidRequestError{err: fmt.Errorf("invalid request body: %w", err)}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return InvalidRequestError{err: errors.New("request body must contain a single JSON value")}
	}
	return nil
}

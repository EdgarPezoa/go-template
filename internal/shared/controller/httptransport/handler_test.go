package httptransport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWritesMappedError(t *testing.T) {
	recorder := httptest.NewRecorder()
	Handler(func(http.ResponseWriter, *http.Request) error {
			return errors.New("bad")
		}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d", recorder.Code)
	}
}

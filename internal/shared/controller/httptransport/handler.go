package httptransport

import (
	"net/http"
)

type Endpoint func(http.ResponseWriter, *http.Request) error

func Handler(endpoint Endpoint) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if err := endpoint(responseWriter, request); err != nil {
			WriteError(responseWriter, evaluate(err))
		}
	})
}

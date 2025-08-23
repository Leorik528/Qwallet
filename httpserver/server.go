package httpserver

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	httpHandlers *HTTPHandlers
}

func NewHTTPServer(httpHandler *HTTPHandlers) *HTTPServer {
	return &HTTPServer{
		httpHandlers: httpHandler,
	}
}

func (s *HTTPServer) StartServer() error {
	router := mux.NewRouter()

	router.Path("/api/send").Methods("POST").HandlerFunc(s.httpHandlers.HandleSend)
	router.Path("/api/wallet/{address}/balance").Methods("GET").HandlerFunc(s.httpHandlers.HandleGetBalance)
	router.Path("/api/transactions").Methods("GET").Queries("count", "{count}").HandlerFunc(s.httpHandlers.HandleGetLast)

	if err := http.ListenAndServe(":8080", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}

	return nil
}

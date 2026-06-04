package devserver

import "net/http"

type Server struct {
	Addr string
}

func New(addr string) *Server {
	return &Server{Addr: addr}
}

func (s *Server) Start() error {
	if s.Addr == "" {
		s.Addr = "127.0.0.1:5173"
	}
	return http.ListenAndServe(s.Addr, http.DefaultServeMux)
}

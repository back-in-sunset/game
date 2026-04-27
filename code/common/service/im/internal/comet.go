package internal

import "im/conf"

type Server struct {
	c     *conf.Config
	round *Round
}

func NewServer(c *conf.Config) *Server {
	return &Server{
		c:     c,
		round: NewRound(c),
	}
}

func (s *Server) onlineproc() {
	for {

	}
}

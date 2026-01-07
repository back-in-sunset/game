package internal

import (
	"im/conf"
	"im/pkg/bufio"
	"im/pkg/bytes"
	"im/pkg/time"
	"log/slog"
	"net"

	log "github.com/golang/glog"
)

const (
	maxInt = 1<<31 - 1
)

func InitTCP(server *Server, addrs []string, accept int) (err error) {
	var (
		bind     string
		addr     *net.TCPAddr
		listener *net.TCPListener
	)

	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			slog.Error("net.ResolveTCPAddr ", "tcp: ", bind, "err: ", err)
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			slog.Error("net.ListenTCP ", "tcp: ", bind, "err: ", err)
			return
		}
		slog.Info("start tcp ", "listen: ", bind)
		for range accept {
			go acceptTCP(server, listener)
		}
	}
	return
}

func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			slog.Error("listener.Accept: ", lis.Addr().String(), err.Error())
		}
		if err = conn.SetKeepAlive(server.c.TCP.KeepAlive); err != nil {
			slog.Error("conn.SetReadBuffer() ", "error: ", err.Error())
			return
		}
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			slog.Error("conn.SetReadBuffer() ", "error: ", err.Error())
			return
		}
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			slog.Error("conn.SetWriteBuffer() ", "error: ", err.Error())
			return
		}

		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(s *Server, conn *net.TCPConn, r int) {
	var (
		tr = s.round.Timer(r)
		rp = s.round.Reader(r)
		wp = s.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if conf.Conf.Debug {
		log.Infof("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)
	}
	s.ServeTCP(conn, rp, wp, tr)
}

func (s *Server) ServeTCP(c *net.TCPConn, rp *bytes.Pool, wp *bytes.Pool, tr *time.Timer) {
	var (
		rb = rp.Get()
		wb = wp.Get()
		rr = new(bufio.Reader)
		wr = new(bufio.Writer)
	)
	rr.ResetBuffer(c, rb.Bytes())
	wr.ResetBuffer(c, wb.Bytes())
	for {
		slog.Info("start recive")
		buf, err := rr.Pop(4)
		if err != nil {
			panic(err)
		}
		slog.Info(string(buf))
		slog.Info("end recive")

	}
}

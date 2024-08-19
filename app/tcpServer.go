package app

import (
	"errors"
	"ezturp/protocol"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	BUF_SIZE                 = 1024 * 8
	INTERNAL_CONN_IDLE       = 1000000000
	KEEP_ALIVE               = INTERNAL_CONN_IDLE / 2
	MAINTAIN_UDP_CLIENT_ADDR = 60
)

type TcpServer struct {
	reacceptSig         chan interface{}
	internalAcceptedSig chan interface{}
	externalConnMutex   sync.Mutex
	internalConnMutex   sync.Mutex
	externalConns       map[uint32]net.Conn
	internalConn        net.Conn
}

func (s *TcpServer) init() {
	s.reacceptSig = make(chan interface{}, 1)
	s.internalAcceptedSig = make(chan interface{}, 1)
	s.externalConns = map[uint32]net.Conn{}
}

func (s *TcpServer) Listen(internalAddr, externalAddr string) error {
	s.init()
	internalListener, err := net.Listen("tcp", internalAddr)
	if err != nil {
		return err
	}
	defer internalListener.Close()
	externalListener, err := net.Listen("tcp", externalAddr)
	if err != nil {
		return err
	}
	defer externalListener.Close()
	go s.listenInternal(internalListener)
	go s.listenExternal(externalListener)
	s.dispatch()
	return nil
}

func (s *TcpServer) listenInternal(listener net.Listener) {
	for {
		log.Printf("listen internal connection %v", listener.Addr())
		internalConn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		s.internalConnMutex.Lock()
		s.internalConn = internalConn
		s.internalConnMutex.Unlock()

		s.internalAcceptedSig <- struct{}{}
		log.Printf("internal %v connected", internalConn.RemoteAddr())
		<-s.reacceptSig

		s.internalConnMutex.Lock()
		_ = internalConn.Close()
		s.internalConn = nil
		s.internalConnMutex.Unlock()

		log.Printf("internal %v disconnected", internalConn.RemoteAddr())
	}
}
func (s *TcpServer) internalNil() bool {
	s.internalConnMutex.Lock()
	defer s.internalConnMutex.Unlock()
	return s.internalConn == nil
}
func (s *TcpServer) listenExternal(listener net.Listener) {
	log.Printf("listen external connection %v", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		if s.internalNil() {
			_ = conn.Close()
			continue
		}
		err, id := s.sessionCreate(conn)
		if err != nil {
			log.Printf("failed to accept external connection %v", err)
			conn.Close()
		}
		go s.proxy(conn, id)
		log.Printf("%v connected", conn.RemoteAddr())
	}
}

func (s *TcpServer) sessionFind(id uint32) (conn net.Conn) {
	s.externalConnMutex.Lock()
	defer s.externalConnMutex.Unlock()
	conn, _ = s.externalConns[id]
	return conn
}

func (s *TcpServer) sessionRemove(id uint32) {
	s.externalConnMutex.Lock()
	defer s.externalConnMutex.Unlock()
	if conn, ok := s.externalConns[id]; ok {
		_ = conn.Close()
		log.Printf("session %v,address %v removed", id, conn.RemoteAddr())
		delete(s.externalConns, id)
		_ = s.internalWriteFrame(protocol.REMOVE_SESSION, id, []byte{})
	}
}

func (s *TcpServer) internalWriteFrame(t byte, id uint32, data []byte) (err error) {
	if !s.internalNil() {
		err = protocol.WriteFrame(s.internalConn, t, id, data)
	} else {
		err = errors.New("internal connection is disabled")
	}
	return err
}

func (s *TcpServer) sessionCreate(conn net.Conn) (error, uint32) {
	var id uint32
	s.externalConnMutex.Lock()
	for {
		id = rand.Uint32()
		if _, ok := s.externalConns[id]; !ok {
			break
		}
	}
	err := s.internalWriteFrame(protocol.NEW_SESSION, id, []byte{})
	if err != nil {
		return err, 0
	}
	s.externalConns[id] = conn
	s.externalConnMutex.Unlock()
	return nil, id
}

func (s *TcpServer) proxy(conn net.Conn, id uint32) {
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("external %v disconnected", conn.RemoteAddr())
			break
		}
		if s.internalConn == nil {
			break
		}
		err = s.internalWriteFrame(protocol.DATA, id, buf[:n])
		if err != nil {
			break
		}
	}
	s.sessionRemove(id)
}

/*
	func (s *TcpServer) reaccept() {
		select {
		case s.reacceptSig <- struct{}{}:
		}
	}
*/

func (s *TcpServer) internalReadFrame(timeout time.Time) (t byte, id uint32, data []byte, err error) {

	if !s.internalNil() {
		err = s.internalConn.SetReadDeadline(timeout)
		if err != nil {
			return
		}
		t, id, data, err = protocol.ReadFrame(s.internalConn)
	} else {
		err = errors.New("internal connection is disabled")
	}
	return
}

func (s *TcpServer) dispatch() {
	for {
		if s.internalNil() {
			<-s.internalAcceptedSig
		}
		t, id, data, err := s.internalReadFrame(time.Now().Add(INTERNAL_CONN_IDLE * time.Second))
		if err != nil {
			s.reacceptSig <- struct{}{}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if t == protocol.KEEP_ALIVE {
			continue
		}
		conn := s.sessionFind(id)
		if conn == nil {
			continue
		}
		s.handleInternalMsg(conn, t, id, data)
	}

}

func (s *TcpServer) handleInternalMsg(conn net.Conn, t byte, id uint32, data []byte) {
	switch t {
	case protocol.DATA:
		_, err := conn.Write(data)
		if err != nil {
			s.sessionRemove(id)
		}
	case protocol.REMOVE_SESSION:
		s.sessionRemove(id)
	default:
		log.Printf("unknown message type : %v in session %v", t, id)
		s.sessionRemove(id)
	}
}

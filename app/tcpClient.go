package app

import (
	"ezturp/protocol"
	"log"
	"net"
	"sync"
	"time"
)

type TcpClient struct {
	LocalAddr    string
	internalConn net.Conn
	sessionMutex sync.Mutex
	sessions     map[uint32]net.Conn
}

func (c *TcpClient) init() {
	c.sessions = make(map[uint32]net.Conn)
}

func (c *TcpClient) Connect(internalAddr string) error {
	c.init()
	conn, err := net.Dial("tcp", internalAddr)
	if err != nil {
		return err
	}
	c.internalConn = conn
	log.Printf("connected to %v", internalAddr)
	go c.keepAlive(conn)
	err = c.handle(conn)
	return err
}

func (c *TcpClient) keepAlive(conn net.Conn) {
	ticker := time.NewTicker(KEEP_ALIVE * time.Second)
	for {
		<-ticker.C
		err := protocol.WriteFrame(conn, protocol.KEEP_ALIVE, 0, []byte{})
		if err != nil {
			break
		}
	}
	ticker.Stop()
	_ = conn.Close()
}

func (c *TcpClient) handle(internalConn net.Conn) (err error) {
	defer internalConn.Close()
	for {
		t, id, data, err := protocol.ReadFrame(internalConn)
		if err != nil {
			break
		}
		switch t {
		case protocol.NEW_SESSION:
			err = c.sessionCreate(id)
			if err != nil {
				log.Printf("failed to create session %v", id)
				c.sessionRemove(id, true)
			}
		case protocol.REMOVE_SESSION:
			c.sessionRemove(id, false)
		case protocol.DATA:
			c.dataDispatch(id, data)
		default:
			log.Printf("unknown message type :%v", t)
		}
	}
	return err
}

func (c *TcpClient) sessionCreate(id uint32) error {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	conn, err := net.Dial("tcp", c.LocalAddr)
	if err != nil {
		return err
	}
	c.sessions[id] = conn
	log.Printf("session %v created", id)
	go c.proxy(conn, id)
	return nil
}

func (c *TcpClient) sessionRemove(id uint32, notify bool) {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	if conn, ok := c.sessions[id]; ok {
		delete(c.sessions, id)
		log.Printf("session %v,address %v removed", id, conn.RemoteAddr())
		_ = conn.Close()
	}
	if notify {
		err := protocol.WriteFrame(c.internalConn, protocol.REMOVE_SESSION, id, []byte{})
		if err != nil {
			log.Printf("failed to notify server to remove session %v : %v", id, err)
		}
	}
}

func (c *TcpClient) proxy(conn net.Conn, id uint32) {
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			c.sessionRemove(id, true)
			log.Printf("local %v disconnected", conn.RemoteAddr())
			break
		}
		err = protocol.WriteFrame(c.internalConn, protocol.DATA, id, buf[:n])
		if err != nil {
			c.sessionRemove(id, false)
			log.Printf("internal connection error : %v", err)
			break
		}
	}
	_ = conn.Close()
}

func (c *TcpClient) dataDispatch(id uint32, data []byte) {
	conn := c.sessionFind(id)
	if conn != nil {
		_, err := conn.Write(data)
		if err != nil {
			log.Printf("local %v disconnected", conn.RemoteAddr())
			c.sessionRemove(id, true)
		}
	} else {
		c.sessionRemove(id, false)
	}
}

func (c *TcpClient) sessionFind(id uint32) net.Conn {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	conn, _ := c.sessions[id]
	return conn
}

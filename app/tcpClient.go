package app

import (
	"ezturp/protocol"
	"ezturp/tools"
	"net"
	"sync"
	"time"
)

type TcpClient struct {
	Name         string
	logger       tools.Logger
	LocalAddr    string
	internalConn net.Conn
	sessionMutex sync.Mutex
	sessions     map[uint32]net.Conn
}

func (c *TcpClient) init() {
	c.sessions = make(map[uint32]net.Conn)
	c.logger = tools.Logger{"TcpClient", c.Name}
}

func (c *TcpClient) Connect(internalAddr string) error {
	c.init()
	conn, err := net.Dial("tcp", internalAddr)
	if err != nil {
		return err
	}
	c.internalConn = conn
	c.logger.Info("connected to %v", internalAddr)
	go c.keepAlive(conn)
	err = c.handle(conn)
	return err
}

func (c *TcpClient) keepAlive(internalConn net.Conn) {
	ticker := time.NewTicker(KEEP_ALIVE)
	for {
		<-ticker.C
		err := protocol.WriteFrame(internalConn, protocol.KEEP_ALIVE, 0, []byte{})
		if err != nil {
			break
		}
	}
	ticker.Stop()
	_ = internalConn.Close()
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
				c.logger.Warn("failed to create session %v", id)
				c.sessionRemove(id, true)
			}
		case protocol.REMOVE_SESSION:
			c.sessionRemove(id, false)
		case protocol.DATA:
			c.dataDispatch(id, data)
		default:
			c.logger.Warn("unknown message type :%v", t)
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
	c.logger.Debug("session %v created", id)
	go c.proxy(conn, id)
	return nil
}

func (c *TcpClient) sessionRemove(id uint32, notify bool) {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	if conn, ok := c.sessions[id]; ok {
		delete(c.sessions, id)
		c.logger.Debug("session %v,address %v removed", id, conn.RemoteAddr())
		err := conn.Close()
		if err != nil {
			c.logger.Warn("connection %v did not close, session id: %v", conn.RemoteAddr().String(), id)
		}
	}
	if notify {
		err := protocol.WriteFrame(c.internalConn, protocol.REMOVE_SESSION, id, []byte{})
		if err != nil {
			c.logger.Warn("failed to notify server to remove session %v : %v", id, err)
		}
	}
}

func (c *TcpClient) proxy(conn net.Conn, id uint32) {
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			c.sessionRemove(id, true)
			c.logger.Info("local %v disconnected", conn.RemoteAddr())
			break
		}
		err = protocol.WriteFrame(c.internalConn, protocol.DATA, id, buf[:n])
		if err != nil {
			c.sessionRemove(id, false)
			c.logger.Error("internal connection error : %v", err)
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

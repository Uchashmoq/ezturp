package app

import (
	"ezturp/protocol"
	"ezturp/tools"
	"net"
	"sync"
	"time"
)

type UdpClient struct {
	Name         string
	logger       tools.Logger
	LocalAddr    string
	localAddr    *net.UDPAddr
	internalConn *net.UDPConn

	sessionMutex      sync.Mutex
	sessionConnMap    map[uint32]*net.UDPConn
	sessionTimeoutMap map[uint32]*time.Timer
}

const (
	UDP_CLIENT_IDLE = 30 * time.Minute
)

func (c *UdpClient) init() {
	c.sessionConnMap = make(map[uint32]*net.UDPConn)
	c.sessionTimeoutMap = make(map[uint32]*time.Timer)
	c.logger = tools.Logger{"UdpClient", c.Name}
}

func (c *UdpClient) Connect(internalAddr string) error {
	c.init()
	localAddr, err := net.ResolveUDPAddr("udp", c.LocalAddr)
	c.localAddr = localAddr
	if err != nil {
		return err
	}
	addr, err := net.ResolveUDPAddr("udp", internalAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	c.internalConn = conn
	go c.maintainClientAddr()
	return c.handleInternal()
}

func (c *UdpClient) maintainClientAddr() {
	ticker := time.NewTicker(MAINTAIN_UDP_CLIENT_ADDR * time.Second)
	for {
		err := protocol.WriteFrame(c.internalConn, protocol.MAINTAIN_UDP_CLIENT_ADDR, 0, []byte{})
		if err != nil {
			break
		}
		<-ticker.C
	}
	ticker.Stop()
	_ = c.internalConn.Close()
}

func (c *UdpClient) handleInternal() (err error) {
	buf := make([]byte, BUF_SIZE)
	c.logger.Info("waiting for message from %v", c.internalConn.RemoteAddr())
	for {
		n, _, err := c.internalConn.ReadFromUDP(buf)
		if err != nil {
			//log.Printf("udp client receiving data error : %v", err)
			c.logger.Error("receiving data error : %v", err)
			break
		}
		t, id, data, err := protocol.ParseFrame(buf[:n])
		if err != nil {
			//log.Printf("udp client parsing frame error : %v", err)
			c.logger.Error("parsing frame error : %v", err)
			break
		}
		switch t {
		case protocol.DATA:
			c.dispatch(id, data)
		default:
			c.logger.Warn("unknown message type : %v in session %v", t, id)
		}
	}
	c.internalConn.Close()
	return
}

func (c *UdpClient) dispatch(id uint32, data []byte) {
	conn, err := c.getConn(id)
	if err != nil {
		c.logger.Warn("getting udp connection error %v", err)
	}
	c.logger.Debug("session %v <- %d bytes", id, len(data))
	c.resetSessionTimeout(id)
	_, err = conn.Write(data)
	if err != nil {
		_ = conn.Close()
	}
}
func (c *UdpClient) resetSessionTimeout(id uint32) {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	tm, ok := c.sessionTimeoutMap[id]
	if ok {
		tm.Reset(UDP_CLIENT_IDLE)
	}
}

func (c *UdpClient) getConn(id uint32) (*net.UDPConn, error) {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	if conn, ok := c.sessionConnMap[id]; ok {
		return conn, nil
	}

	newConn, err := net.DialUDP("udp", nil, c.localAddr)
	if err != nil {
		return nil, err
	}
	c.sessionConnMap[id] = newConn
	c.sessionTimeoutMap[id] = time.AfterFunc(UDP_CLIENT_IDLE, func() {
		c.removeSession(id)
		c.logger.Info("session %v is idle", id)
	})
	go c.proxy(newConn, id)
	c.logger.Info("created a new session %v,address %v", id, newConn.LocalAddr().String())
	return newConn, nil
}

func (c *UdpClient) proxy(newConn *net.UDPConn, id uint32) {
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := newConn.Read(buf)
		c.resetSessionTimeout(id)
		if err != nil {
			//c.logger.Error("receiving data from server error :%v", err)
			break
		}
		err = protocol.WriteFrame(c.internalConn, protocol.DATA, id, buf[:n])
		if err != nil {
			c.logger.Error("sending data to server error :%v", err)
			break
		}
		c.logger.Debug("session %v , %v sent %v to server", id, newConn.LocalAddr(), n)
	}
	_ = newConn.Close()
}

func (c *UdpClient) removeSession(id uint32) {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()
	conn, ok := c.sessionConnMap[id]
	if ok {
		delete(c.sessionConnMap, id)
		tm, ok1 := c.sessionTimeoutMap[id]
		if ok1 {
			tm.Stop()
			delete(c.sessionTimeoutMap, id)
		}
		conn.Close()
	}
}

package app

import (
	"ezturp/protocol"
	"log"
	"net"
	"sync"
	"time"
)

type UdpClient struct {
	LocalAddr    string
	localAddr    *net.UDPAddr
	internalConn *net.UDPConn

	sessionMutex   sync.Mutex
	sessionConnMap map[uint32]*net.UDPConn
	connSessionMap map[*net.UDPConn]uint32
}

func (c *UdpClient) init() {
	c.sessionConnMap = make(map[uint32]*net.UDPConn)
	c.connSessionMap = make(map[*net.UDPConn]uint32)
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
	for {
		n, _, err := c.internalConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("udp client receiving data error : %v", err)
			break
		}
		t, id, data, err := protocol.ParseFrame(buf[:n])
		if err != nil {
			log.Printf("udp client parsing frame error : %v", err)
			break
		}
		switch t {
		case protocol.DATA:
			c.dispatch(id, data)
		default:
			log.Printf("unknown message type : %v in session %v", t, id)
		}
	}
	c.internalConn.Close()
	return
}

func (c *UdpClient) dispatch(id uint32, data []byte) {
	conn, err := c.getConn(id)
	if err != nil {
		log.Printf("getting udp connection error %v", err)
	}
	_, err = conn.Write(data)
	if err != nil {
		_ = conn.Close()
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
	go c.proxy(newConn, id)
	return newConn, nil
}

func (c *UdpClient) proxy(newConn *net.UDPConn, id uint32) {
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := newConn.Read(buf)
		if err != nil {
			break
		}
		err = protocol.WriteFrame(c.internalConn, protocol.DATA, id, buf[:n])
		if err != nil {
			log.Printf("udp client sending data to server error :%v", err)
		}
	}
	_ = newConn.Close()
}

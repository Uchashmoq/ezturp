package app

import (
	"ezturp/protocol"
	"log"
	"math/rand"
	"net"
	"sync"
)

type UdpServer struct {
	clientAddr      *net.UDPAddr
	clientAddrMutex sync.Mutex
	internalConn    *net.UDPConn
	externalConn    *net.UDPConn

	addrSessionMap map[string]uint32
	sessionAddrMap map[uint32]string
	sessionMutex   sync.Mutex
}

func (s *UdpServer) init() {
	s.addrSessionMap = make(map[string]uint32)
	s.sessionAddrMap = make(map[uint32]string)
}

func (s *UdpServer) Listen(internalAddr, externalAddr string) error {
	s.init()
	internalUdpAddr, err := net.ResolveUDPAddr("udp", internalAddr)
	if err != nil {
		return err
	}
	externalUdpAddr, err := net.ResolveUDPAddr("udp", externalAddr)
	if err != nil {
		return err
	}
	err = s.listenInternal(internalUdpAddr)
	if err != nil {
		return err
	}
	go s.handleInternalMsg()
	err = s.listenExternal(externalUdpAddr)
	s.recvExternalMsg()
	return nil
}

func (s *UdpServer) handleExternalMsg(data []byte, addr *net.UDPAddr) {
	addrStr := addr.String()
	id := s.getSessionId(addrStr)
	frame := protocol.EncodeFrame(protocol.DATA, id, data)
	s.clientAddrMutex.Lock()
	s.internalConn.WriteToUDP(frame, s.clientAddr)
	s.clientAddrMutex.Unlock()
}

func (s *UdpServer) recvExternalMsg() {
	buf := make([]byte, BUF_SIZE)
	for {
		n, remoteAddr, err := s.externalConn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		s.handleExternalMsg(buf[:n], remoteAddr)
	}
}

func (s *UdpServer) listenExternal(addr *net.UDPAddr) error {
	externalConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("udp server failed to listen external connection at %v:%v", *addr, err)
		return err
	}
	log.Printf("udp server listening external connection at %v", *addr)
	s.externalConn = externalConn
	return nil
}

func (s *UdpServer) listenInternal(addr *net.UDPAddr) error {
	internalConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("udp server failed to listen internal connection at %v:%v", *addr, err)
		return err
	}
	log.Printf("udp server listening internal connection at %v", *addr)
	s.internalConn = internalConn
	return nil
}

func (s *UdpServer) handleInternalMsg() {
	buf := make([]byte, BUF_SIZE)
	for {
		n, clientAddr, err := s.internalConn.ReadFromUDP(buf)
		t, id, data, err := protocol.ParseFrame(buf[:n])
		if err != nil {
			log.Printf("error in handling internal message : %v", err)
			continue
		}
		switch t {
		case protocol.MAINTAIN_UDP_CLIENT_ADDR:
			s.setClientAddr(clientAddr)
		case protocol.DATA:
			s.dispatch(id, data)
		}

	}

}

func (s *UdpServer) setClientAddr(addr *net.UDPAddr) {
	s.clientAddrMutex.Lock()
	s.clientAddr = addr
	s.clientAddrMutex.Unlock()
}

func (s *UdpServer) getSessionId(addrStr string) uint32 {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	id, ok := s.addrSessionMap[addrStr]
	if ok {
		return id
	}
	var newId uint32
	for {
		newId := rand.Uint32()
		if _, ok := s.sessionAddrMap[newId]; !ok {
			break
		}
	}
	s.sessionAddrMap[newId] = addrStr
	s.addrSessionMap[addrStr] = newId
	return newId
}

func (s *UdpServer) dispatch(id uint32, data []byte) {
	addr, ok := s.getAddr(id)
	if !ok {
		log.Printf("in udp server, unkonwn session id %v", id)
	}
	_, err := s.externalConn.WriteToUDP(data, addr)
	if err != nil {
		log.Printf("udp server %v", err)
	}
}

func (s *UdpServer) getAddr(id uint32) (addr *net.UDPAddr, ok bool) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	addrStr, ok := s.sessionAddrMap[id]
	if !ok {
		return
	}
	addr, _ = net.ResolveUDPAddr("udp", addrStr)
	return
}

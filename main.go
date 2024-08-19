package main

import (
	"ezturp/app"
	"log"
	"os"
	"time"
)

func main() {
	//testServer1()
	//testTcpServer()
	//testTcpClient()
	//testCS()
	testUdpCS()
}

func testServer1() {
	s := app.TcpServer{}
	if len(os.Args) < 3 {
		log.Fatalln("internal external")
	}
	err := s.Listen(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
}

func testCS() {
	go testTcpServer()
	time.Sleep(100 * time.Millisecond)
	testTcpClient()
}

func testTcpServer() {
	s := app.TcpServer{}
	err := s.Listen("127.0.0.1:10001", "127.0.0.1:10002")
	if err != nil {
		log.Fatalln(err)
	}
}
func testTcpClient() {
	c := app.TcpClient{LocalAddr: "127.0.0.1:10009"}
	err := c.Connect("47.108.118.112:23891")
	//err := c.Connect("127.0.0.1:10001")
	if err != nil {
		log.Fatalln(err)
	}
}

func testUdpServer() {
	s := app.UdpServer{}
	err := s.Listen("127.0.0.1:20001", "127.0.0.1:20002")
	if err != nil {
		log.Fatalln(err)
	}
}

func testUdpClient() {
	c := app.UdpClient{LocalAddr: "127.0.0.1:20000"}
	err := c.Connect("127.0.0.1:20001")
	if err != nil {
		log.Fatalln(err)
	}
}

func testUdpCS() {
	go testUdpServer()
	time.Sleep(1 * time.Second)
	testUdpClient()
}

package main

import (
	"ezturp/app"
	"ezturp/tools"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	OP_TCP_SERVER     = "ts"
	OP_UDP_SERVER     = "us"
	OP_TCP_CLIENT     = "tc"
	OP_UDP_CLIENT     = "uc"
	OP_CLIENT_MANAGER = "cm"
	OP_SERVER_MANAGER = "sm"
	OP_JSON           = "json"
	OP_CONFIG         = "config"
	OP_INTERNAL_ADDR  = "iaddr"
	OP_EXTERNAL_ADDR  = "eaddr"
	OP_LOCAL_ADDR     = "laddr"
	OP_NAME           = "n"
	OP_LOG            = "log"
)

func main() {
	//testServer1()
	//testTcpServer()
	//testTcpClient()
	//testCS()
	//testUdpCS()
	args := tools.ParseCommandArgs(os.Args)
	if args.ContainsOpt(OP_LOG) {
		initLogger(args)
	}
	switch {
	case args.ContainsOpt(OP_TCP_SERVER):
		launchTcpServer(args)
	case args.ContainsOpt(OP_UDP_SERVER):
		launchUdpServer(args)
	case args.ContainsOpt(OP_TCP_CLIENT):
		launchTcpClient(args)
	case args.ContainsOpt(OP_UDP_CLIENT):
		launchUdpClient(args)
	case args.ContainsOpt(OP_SERVER_MANAGER):
		launchServerManager(args)
	case args.ContainsOpt(OP_CLIENT_MANAGER):
		launchClientManager(args)
	default:
		panic(fmt.Errorf("\n%s <-%v | -%v | -%v | -%v> <-%v | -%v | -%v | -%v | -%v> [-%v] [-%v [level] [path]]",
			os.Args[0], OP_TCP_SERVER, OP_UDP_SERVER, OP_TCP_CLIENT, OP_UDP_CLIENT,
			OP_INTERNAL_ADDR, OP_EXTERNAL_ADDR, OP_LOCAL_ADDR,
			OP_JSON, OP_CONFIG, OP_NAME,
			OP_LOG,
		))
	}
}

func initLogger(args tools.CommandArgs) {
	tools.SetLevelStr(args.GetDefault(OP_LOG, 0, ""))
	tools.SetLogOutput(args.GetDefault(OP_LOG, 1, ""))
}

func launchClientManager(args tools.CommandArgs) {
	var json []byte
	if args.ContainsOpt(OP_JSON) {
		json = []byte(args.Get0(OP_JSON))
	} else {
		var err error
		json, err = os.ReadFile(args.Get0(OP_CONFIG))
		if err != nil {
			panic(err)
		}
	}
	app.StartClientManager(
		args.Get0Default(OP_NAME, ""),
		app.LoadClientConfigsFromJson(json),
	)
	select {}
}

func launchServerManager(args tools.CommandArgs) {
	var json []byte
	if args.ContainsOpt(OP_JSON) {
		json = []byte(args.Get0(OP_JSON))
	} else {
		var err error
		json, err = os.ReadFile(args.Get0(OP_CONFIG))
		if err != nil {
			panic(err)
		}
	}
	app.StartServerManager(
		args.Get0Default(OP_NAME, ""),
		app.LoadServerConfigsFromJson(json),
	)
	select {}
}

func launchTcpClient(args tools.CommandArgs) {
	c := app.TcpClient{Name: args.Get0Default(OP_NAME, ""), LocalAddr: args.Get0(OP_LOCAL_ADDR)}
	err := c.Connect(args.Get0(OP_INTERNAL_ADDR))
	if err != nil {
		fmt.Println(err)
	}
}

func launchUdpClient(args tools.CommandArgs) {
	c := app.UdpClient{Name: args.Get0Default(OP_NAME, ""), LocalAddr: args.Get0(OP_LOCAL_ADDR)}
	err := c.Connect(args.Get0(OP_INTERNAL_ADDR))
	if err != nil {
		fmt.Println(err)
	}
}

func launchTcpServer(args tools.CommandArgs) {
	s := app.TcpServer{Name: args.Get0Default(OP_NAME, "")}
	err := s.Listen(args.Get0(OP_INTERNAL_ADDR), args.Get0(OP_EXTERNAL_ADDR))
	if err != nil {
		fmt.Println(err)
	}
}
func launchUdpServer(args tools.CommandArgs) {
	s := app.UdpServer{Name: args.Get0Default(OP_NAME, "")}
	err := s.Listen(args.Get0(OP_INTERNAL_ADDR), args.Get0(OP_EXTERNAL_ADDR))
	if err != nil {
		fmt.Println(err)
	}
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

# Eazy Tcp Udp Reverse Proxy

## TcpClient	UdpClient	TcpServer 	Udpserver

### Using in the code

```go
type TcpServer struct {
	Name                string //optional
	logger              tools.Logger
	reacceptSig         chan interface{}
	internalAcceptedSig chan interface{}
	externalConnMutex   sync.Mutex
	internalConnMutex   sync.Mutex
	externalConns       map[uint32]net.Conn
	internalConn        net.Conn
}

func (s *TcpServer) Listen(internalAddr, externalAddr string) error

type TcpClient struct {
	Name         string
    LocalAddr    string
	logger       tools.Logger
	internalConn net.Conn
	sessionMutex sync.Mutex
	sessions     map[uint32]net.Conn
}

func (c *TcpClient) Connect(internalAddr string) error
```

The client connects to the server using the `internalAddr`, and through this connection, the client and server communicate. The `externalAddr` is bound to the public IP address, allowing external connections to reach the server. When an external connection is made to the server, the server uses the `internalConn` to control the client's access to the local address `LocalAddr`, thereby establishing a proxy tunnel.

example:

```go
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
	err := c.Connect("48.107.117.113:23891")
	//err := c.Connect("127.0.0.1:10001")
	if err != nil {
		log.Fatalln(err)
	}
}
```

`UdpClient` ,`UdpServer` have the same usage.

### Starting with commands

client:

```bash
./ezturp 
-tc #client type: tc(TcpClient) , uc(UdpClient)
-laddr "127.0.0.1:80" # local address
-iaddr "48.107.117.113:23891" #proxy server address
-n "httpProxy" #name (optional)
-log "debug" "logs/clientLog.txt" #(log optional)
```

server

```bash
./ezturp
-ts #server type: ts(TcpServer) , us(UdpServer)
-eaddr "0.0.0.0:8080" #external address,external devices will access this address
-iaddr "48.107.117.113:23891" #proxy server address
-n "httpProxy" #name (optional)
-log "debug" "logs/clientLog.txt" #log （optional)
```

## ClientManager

The `ClientManager` is a component designed to manage multiple network clients, handling both TCP and UDP connections. Its primary function is to initialize and manage these clients based on a given configuration, ensuring that they remain operational even if they encounter errors. The `ClientManager` automatically restarts clients in case of failures, allowing for resilient and continuous network communication.

### Using in the code

1. **Load Configurations**:

   - Use `LoadClientConfigsFromJson(p []byte)` to load client configurations from a JSON file or byte slice. This function returns a slice of `ClientConfig` structures, each containing details for a specific client.

   ```go
   type ClientConfig struct {
   	Name            string `json:"name"`
   	Protocol        string `json:"protocol"`
   	LocalAddress    string `json:"local_address"`
   	InternalAddress string `json:"internal_address"`
   }
   
   func LoadClientConfigsFromJson(p []byte) []*ClientConfig {
   	var configs []*ClientConfig
   	err := json.Unmarshal(p, &configs)
   	if err != nil {
   		panic(err)
   	}
   	return configs
   }
   ```

   example of configuration file:

   ```json
   [
   {
       "name": "sunshineUdp1",
       "protocol": "udp",
       "local_address": "127.0.0.1:47998",
       "internal_address": "127.0.0.1:48998"
     },
     {
       "name": "sunshineUdp2",
       "protocol": "udp",
       "local_address": "127.0.0.1:47999",
       "internal_address": "127.0.0.1:48999"
     }
     ]
   ```

   

2. **Start the Manager**:

   - Call `StartClientManager(name string, configs []*ClientConfig)` to create and start the `ClientManager`. Pass in a name for the manager and the slice of client configurations. The manager will automatically start the appropriate clients based on their protocol.

   ```go
   func StartClientManager(name string, configs []*ClientConfig)
   select{} //Preventing program exit
   ```

   

3. **Automatic Operation**:

   - Once started, the `ClientManager` will handle all client operations automatically, including managing errors and restarting clients as needed. You can monitor its operations through the log messages.

### Starting with commands

```bash
#Starting via configuration file
./ezturp
-cm #manager type: cm(ClientManager) sm(ServerManager)
-config "configs/clientConfig.json"
-n "test1" #name (optional)
-log "info" "logs/clientLog.txt" #(log optional)


#Launch via json
./ezturp
-cm #manager type: cm(ClientManager) sm(ServerManager)
-json "[{\"name\":\"sunshineTcp2\",\"protocol\":\"tcp\",\"external_address\":\"192.168.0.107:50010\",\"internal_address\":\"127.0.0.1:49010\"},{\"name\":\"sunshineUdp1\",\"protocol\":\"udp\",\"external_address\":\"192.168.0.107:49998\",\"internal_address\":\"127.0.0.1:48998\"}]"

-n "test1" #name (optional)
-log "info" "logs/clientLog.txt" #(log optional)

```



### **ServerManager**

Similarly , the `ServerManager` is responsible for managing multiple network servers, supporting both TCP and UDP protocols. It initializes and manages servers based on the provided configurations, ensuring that they remain operational. If a server encounters an error, the `ServerManager` will automatically restart it, ensuring continuous service availability.

### **Using in the Code**

1. **Load Configurations**:

   - Use `LoadServerConfigsFromJson(p []byte)` to load server configurations from a JSON file or byte slice. This function returns a slice of `ServerConfig` structures, each containing details for a specific server.

   ```go
   type ServerConfig struct {
   	Name            string `json:"name"`
   	Protocol        string `json:"protocol"`
   	InternalAddress string `json:"internal_address"`
   	ExternalAddress string `json:"external_address"`
   }
   
   func LoadServerConfigsFromJson(p []byte) []*ServerConfig {
   	var configs []*ServerConfig
   	err := json.Unmarshal(p, &configs)
   	if err != nil {
   		panic(err)
   	}
   	return configs
   }
   ```

   Example of a configuration file:

   ```json
   [
     {
       "name": "myTcpServer",
       "protocol": "tcp",
       "internal_address": "127.0.0.1:49010",
       "external_address": "192.168.0.107:50010"
     },
     {
       "name": "myUdpServer",
       "protocol": "udp",
       "internal_address": "127.0.0.1:48998",
       "external_address": "192.168.0.107:49998"
     }
   ]
   ```

2. **Start the Manager**:

   - Call `StartServerManager(name string, configs []*ServerConfig)` to create and start the `ServerManager`. Pass in a name for the manager and the slice of server configurations. The manager will automatically start the appropriate servers based on their protocol.

   ```go
   func StartServerManager(name string, configs []*ServerConfig)
   select{} //Preventing program exit
   ```

3. **Automatic Operation**:

   - The `ServerManager` handles all server operations automatically. If a server fails, it logs the error and restarts the server to maintain service continuity.

### **Starting with Commands**

```bash
# Starting via configuration file
./ezturp
-sm # manager type: sm(ServerManager)
-config "configs/serverConfig.json"
-n "testServer" # name (optional)
-log "info" "logs/serverLog.txt" # (log optional)

# Launch via JSON
./ezturp
-sm # manager type: sm(ServerManager)
-json "[{\"name\":\"myTcpServer\",\"protocol\":\"tcp\",\"internal_address\":\"127.0.0.1:49010\",\"external_address\":\"192.168.0.107:50010\"},{\"name\":\"myUdpServer\",\"protocol\":\"udp\",\"internal_address\":\"127.0.0.1:48998\",\"external_address\":\"192.168.0.107:49998\"}]"
-n "testServer" # name (optional)
-log "info" "logs/serverLog.txt" # (log optional)
```

## More examples

### code

```go
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
```

### commands

```bash
./ezturp -tc -laddr ":20000" -iaddr ":20002" -n "test"
./ezturp -ts -iaddr ":20001" -eaddr ":20003" -n "test"
./ezturp -cm -config examples/clientConfig.json
./ezturp -cm -config examples/sunshineClient.json
./ezturp -sm -config examples/moonlightServer.json
```

### configuration files

```json
[
  {
    "name": "sunshineTcp1",
    "protocol": "tcp",
    "external_address": "192.168.0.107:49984",
    "internal_address": "127.0.0.1:48984"
  },
  {
    "name": "sunshineConn",
    "protocol": "tcp",
    "external_address": "192.168.0.107:49989",
    "internal_address": "127.0.0.1:48989"
  },
  {
    "name": "sunshineWebUI",
    "protocol": "tcp",
    "external_address": "192.168.0.107:49990",
    "internal_address": "127.0.0.1:48990"
  },
  {
    "name": "sunshineTcp2",
    "protocol": "tcp",
    "external_address": "192.168.0.107:50010",
    "internal_address": "127.0.0.1:49010"
  },
  {
    "name": "sunshineUdp1",
    "protocol": "udp",
    "external_address": "192.168.0.107:49998",
    "internal_address": "127.0.0.1:48998"
  },
  {
    "name": "sunshineUdp2",
    "protocol": "udp",
    "external_address": "192.168.0.107:49999",
    "internal_address": "127.0.0.1:48999"
  },
  {
    "name": "sunshineUdp3",
    "protocol": "udp",
    "external_address": "192.168.0.107:50000",
    "internal_address": "127.0.0.1:49000"
  }
]

```

```json
[
  {
    "name": "sunshineTcp1",
    "protocol": "tcp",
    "local_address": "127.0.0.1:47984",
    "internal_address": "127.0.0.1:48984"
  },
  {
    "name": "sunshineConn",
    "protocol": "tcp",
    "local_address": "127.0.0.1:47989",
    "internal_address": "127.0.0.1:48989"
  },
  {
    "name": "sunshineWebUI",
    "protocol": "tcp",
    "local_address": "127.0.0.1:47990",
    "internal_address": "127.0.0.1:48990"
  },
  {
    "name": "sunshineTcp2",
    "protocol": "tcp",
    "local_address": "127.0.0.1:48010",
    "internal_address": "127.0.0.1:49010"
  },
  {
    "name": "sunshineUdp1",
    "protocol": "udp",
    "local_address": "127.0.0.1:47998",
    "internal_address": "127.0.0.1:48998"
  },
  {
    "name": "sunshineUdp2",
    "protocol": "udp",
    "local_address": "127.0.0.1:47999",
    "internal_address": "127.0.0.1:48999"
  },
  {
    "name": "sunshineUdp3",
    "protocol": "udp",
    "local_address": "127.0.0.1:48000",
    "internal_address": "127.0.0.1:49000"
  }
]

```


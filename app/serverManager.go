package app

import (
	"encoding/json"
	"ezturp/tools"
)

type ServerConfig struct {
	Name            string `json:"name"`
	Protocol        string `json:"protocol"`
	InternalAddress string `json:"internal_address"`
	ExternalAddress string `json:"external_address"`
}

type ServerManager struct {
	logger tools.Logger
}

func LoadServerConfigsFromJson(p []byte) []*ServerConfig {
	var configs []*ServerConfig
	err := json.Unmarshal(p, &configs)
	if err != nil {
		panic(err)
	}
	return configs
}

func StartServerManager(name string, configs []*ServerConfig) *ServerManager {
	cm := &ServerManager{logger: tools.Logger{
		Service: "ServerManager",
		Name:    name,
	}}
	var cnt int
	for _, cfg := range configs {
		switch cfg.Protocol {
		case UDP:
			go cm.runUdpServer(cfg)
		case TCP:
			go cm.runTcpServer(cfg)
		default:
			cm.logger.Error("unsupported protocol %s", cfg.Protocol)
			panic(cfg.Protocol)
		}
	}
	cm.logger.Info("server manager started , %d servers running", cnt)
	return cm
}

func (cm *ServerManager) runUdpServer(config *ServerConfig) {
	for {
		s := &UdpServer{Name: config.Name}
		err := s.Listen(config.InternalAddress, config.ExternalAddress)
		if err != nil {
			cm.logger.Error("udp server %v error: %v ", config.Name, err)
		}
		cm.logger.Info("udp server %v restart", config.Name)
	}
}
func (cm *ServerManager) runTcpServer(config *ServerConfig) {
	for {
		s := &TcpServer{Name: config.Name}
		err := s.Listen(config.InternalAddress, config.ExternalAddress)
		if err != nil {
			cm.logger.Error("tcp server %v error: %v", config.Name, err)
		}
		cm.logger.Info("tcp server %v restart", config.Name)
	}
}

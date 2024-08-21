package app

import (
	"encoding/json"
	"ezturp/tools"
)

const (
	TCP = "tcp"
	UDP = "udp"
)

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

type ClientManager struct {
	logger tools.Logger
}

func StartClientManager(name string, configs []*ClientConfig) *ClientManager {
	cm := &ClientManager{logger: tools.Logger{
		Service: "ClientManager",
		Name:    name,
	}}
	var cnt int
	for _, cfg := range configs {
		switch cfg.Protocol {
		case UDP:
			go cm.runUdpClient(cfg)
		case TCP:
			go cm.runTcpClient(cfg)
		default:
			cm.logger.Error("unsupported protocol %s", cfg.Protocol)
			panic(cfg.Protocol)
		}
		cnt++
	}
	cm.logger.Info("client manager started , %d clients running", cnt)
	return cm
}

func (cm *ClientManager) runUdpClient(config *ClientConfig) {
	for {
		c := &UdpClient{Name: config.Name, LocalAddr: config.LocalAddress}
		err := c.Connect(config.InternalAddress)
		if err != nil {
			cm.logger.Error("udp client %v error: %v ", config.Name, err)
		}
		cm.logger.Info("udp client %v restart", config.Name)
	}
}
func (cm *ClientManager) runTcpClient(config *ClientConfig) {
	for {
		c := &TcpClient{Name: config.Name, LocalAddr: config.LocalAddress}
		err := c.Connect(config.InternalAddress)
		if err != nil {
			cm.logger.Error("tcp client %v error: %v", config.Name, err)
		}
		cm.logger.Info("tcp client %v restart", config.Name)
	}
}

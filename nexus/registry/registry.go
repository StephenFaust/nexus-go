package registry

type Registry interface {
	ServerRegistry(ServerMataInfo) error
	ServerDiscover(serverName string) (ServerMataInfo, error)
}

type ServerMataInfo struct {
	serverIp   string
	serverPort int
	serverName string
}

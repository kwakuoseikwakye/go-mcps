package mcp

var servers = make(map[string]Server)

// RegisterServer registers an MCP server.
func RegisterServer(s Server) {
	servers[s.Name()] = s
}

// GetServer retrieves a registered server by name.
func GetServer(name string) (Server, bool) {
	s, ok := servers[name]
	return s, ok
}

// ListServers returns the names of all registered servers.
func ListServers() []string {
	keys := make([]string, 0, len(servers))
	for k := range servers {
		keys = append(keys, k)
	}
	return keys
}
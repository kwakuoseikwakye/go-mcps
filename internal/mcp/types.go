package mcp

// Message represents a generic message in an MCP context.
type Message struct {
	Context string
	User    string
	Text    string
	Time    string
}

// Server is a generic MCP server interface.
type Server interface {
	Name() string
	Connect(config map[string]string) error
	ListContexts() ([]string, error)
	SendMessage(ctx, msg string) error
	ReceiveMessage(ctx string) (<-chan Message, error)
}
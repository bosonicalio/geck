package http

// ServerConfig is a configuration structure with general properties for HTTP servers.
//
// Embed this structure to extend its properties when a finer-grained configuration is required
// for a specific driver (e.g. `labstack/echo`).
type ServerConfig struct {
	Address        string `env:"HTTP_SERVER_ADDRESS" envDefault:":8080"`
	ResponseFormat string `env:"HTTP_SERVER_RESPONSE_FORMAT" envDefault:"json"`
}

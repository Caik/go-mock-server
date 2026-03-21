package config

type AppArguments struct {
	MocksDirectory       string `arg:"required,--mocks-directory" help:"path to the mocks directory"`
	MocksConfigFile      string `arg:"--mocks-config-file" help:"path to the config file"`
	DefaultContentType   string `default:"text/plain" arg:"--default-content-type" help:"use default content type when no content type is specified in the request"`
	ServerPort           int    `default:"8080" arg:"-P,--port" help:"port for mock traffic"`
	AdminPort            int    `default:"9090" arg:"--admin-port" help:"port for admin API and UI (0 to disable)"`
	TrafficLogBufferSize int    `default:"1000" arg:"--traffic-log-buffer-size" help:"size of in-memory traffic log buffer (0 to disable traffic logging)"`
	DisableCache         bool   `arg:"--disable-cache" help:"disable the caching"`
	DisableLatency       bool   `arg:"--disable-latency" help:"disable latency simulation"`
DisableCors          bool   `arg:"--disable-cors" help:"disable CORS headers"`
	UIDirectory          string `arg:"--ui-dir" help:"path to the web UI directory to serve"`
}

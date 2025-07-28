package config

type AppArguments struct {
	MocksDirectory  string `arg:"required,--mocks-directory" help:"path to the mocks directory"`
	MocksConfigFile string `arg:"--mocks-config-file" help:"path to the config file"`
	ServerPort      int    `default:"8080" arg:"-P,--port" help:"port to be used"`
	DisableCache    bool   `arg:"--disable-cache" help:"disable the caching"`
	DisableLatency  bool   `arg:"--disable-latency" help:"disable latency simulation"`
	DisableError    bool   `arg:"--disable-error" help:"disable error simulation"`
	DisableCors     bool   `arg:"--disable-cors" help:"disable CORS headers"`
}

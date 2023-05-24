package config

type Config struct {
	ServAddr string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL  string `env:"BASE_URL" envDefault:"http://localhost:8080"`

	FStorPath string `env:"FILE_STORAGE_PATH" envDefault:"../URLs" envExpand:"true"`
}

func (cfg *Config) UpdateByFlags(ServAddr, BaseURL, FStorPath *string) (bool, error) {
	changed := false
	if *BaseURL != "http://localhost:8080" {
		changed = true
		cfg.BaseURL = *BaseURL
	}
	if *ServAddr != ":8080" {
		changed = true
		cfg.ServAddr = *ServAddr
	}
	if *FStorPath != "../URLs" {
		changed = true
		cfg.FStorPath = *FStorPath
	}
	return changed, nil

}

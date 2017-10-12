package config

type Database struct {
	Type             string
	ConnectionString string `yaml:"connection_string"`
	Username         string
	Password         string
	Address          string
	DBName           string
}

type Configuration struct {
	Database Database
	Logs     []LogTarget `yaml:"logs"`
	Jwt      JwtConfig
}

type LogTarget struct {
	Name             string
	Type             string
	Minlevel         string `yaml:"min_level"`
	ConnectionString string `yaml:"connection_string"`
}

type JwtConfig struct {
	SecretKey     string `yaml:"secret_key"`
	DurationInMin int    `yaml:"duration_in_min"`
}

func newConfiguration() *Configuration {
	return &Configuration{
		Database: Database{
			Type: "mysql",
		},
		Jwt: JwtConfig{
			SecretKey:     "",
			DurationInMin: 60,
		},
	}
}

func (c *Configuration) isValid() error {
	return nil
}

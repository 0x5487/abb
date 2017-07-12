package config

type Database struct {
	Username string
	Password string
	Address  string
	DBName   string
}

type Configuration struct {
	Database Database
	Logs     []LogTarget `yaml:"logs"`
}

type LogTarget struct {
	Name             string
	Type             string
	Minlevel         string `yaml:"min_level"`
	ConnectionString string `yaml:"connection_string"`
}

func newConfiguration() *Configuration {
	return &Configuration{}
}

func (c *Configuration) isValid() error {
	return nil
}

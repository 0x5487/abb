package config

import (
	"flag"
	"io/ioutil"
	stdlog "log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/jasonsoft/log"
	"github.com/jasonsoft/log/handlers/console"
)

var (
	_config *Configuration
)

func init() {
	flag.Parse()

	//read and parse config file
	var err error
	rootDirPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		stdlog.Fatalf("config: file error: %s", err.Error())
	}

	configPath := filepath.Join(rootDirPath, "app.yml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// config exists
		file, err := ioutil.ReadFile(configPath)
		if err != nil {
			stdlog.Fatalf("config: file error: %s", err.Error())
		}

		_config = newConfiguration()
		err = yaml.Unmarshal(file, &_config)
		if err != nil {
			stdlog.Fatal("config: config error:", err)
		}
		err = _config.isValid()
		if err != nil {
			stdlog.Fatal(err)
		}
	} else {
		// config does not exist
		_config = &Configuration{
			Database: Database{
				Username: os.Getenv("ABB_DB_USERNAME"),
				Password: os.Getenv("ABB_DB_PASSWORD"),
				Address:  os.Getenv("ABB_DB_ADDRESS"),
				DBName:   os.Getenv("ABB_DB_DBNAME"),
			},
		}
	}

	// set up log target
	for _, target := range _config.Logs {
		switch target.Type {
		case "console":
			clog := console.New()
			levels := log.GetLevelsFromMinLevel(target.Minlevel)
			log.RegisterHandler(clog, levels...)
		case "gelf":
			// graylog := gelf.New(target.ConnectionString)
			// levels := log.GetLevelsFromMinLevel(target.Minlevel)
			// log.RegisterHandler(graylog, levels...)
		}
	}

}

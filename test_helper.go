package slack

import (
	config "github.com/tommzn/go-config"
)

// LoadConfigForTest returns config obtained from passed file. If no file is passed, default is: ixtures/testconfig.yml.
func loadConfigForTest(fileName *string) config.Config {

	configFile := "fixtures/testconfig.yml"
	if fileName != nil {
		configFile = *fileName
	}
	configLoader := config.NewFileConfigSource(&configFile)
	config, _ := configLoader.Load()
	return config
}

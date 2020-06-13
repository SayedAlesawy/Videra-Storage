package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/SayedAlesawy/Videra-Ingestion/orchestrator/utils/errors"
	"gopkg.in/yaml.v2"
)

// logPrefix Used for hierarchical logging
var logPrefix = "[Configuration-Manager]"

// configManagerOnce Used to garauntee thread safety for singleton instances
var configManagerOnce sync.Once

// monitorInstance A singleton instance of the config manager object
var configManagerInstance *ConfigurationManager

// ConfigurationManagerInstance A function to return a configuration manager instance
func ConfigurationManagerInstance(configFilesDir string) *ConfigurationManager {
	configManagerOnce.Do(func() {
		manager := ConfigurationManager{configFilesDir: configFilesDir}

		configManagerInstance = &manager
	})

	return configManagerInstance
}

// retrieveConfig A function to read a config file
func (manager *ConfigurationManager) retrieveConfig(configObj interface{}, filePath string) {
	configFileContent, err := ioutil.ReadFile(filePath)
	errors.HandleError(err, fmt.Sprintf("%s %s\n", logPrefix, fmt.Sprintf("%s %s", "Unable to read config file:", filePath)), true)

	err = yaml.Unmarshal([]byte(configFileContent), configObj)
	errors.HandleError(err, fmt.Sprintf("%s %s\n", logPrefix, fmt.Sprintf("%s %s", "Unable to unmarshal config file:", filePath)), true)
}

// getFilePath A function to get the file path given the name
func (manager *ConfigurationManager) getFilePath(filename string) string {
	filePath := fmt.Sprintf("%s%s", os.ExpandEnv(fmt.Sprintf("%s/", manager.configFilesDir)), filename)

	return filePath
}

// envString Reads env vars with string values
func envString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}

	return defaultValue
}

// envString Reads env vars with integer values
func envInt(key, defaultValue string) int64 {
	envVarValue := envString(key, defaultValue)

	value, err := strconv.ParseInt(envVarValue, 10, 64)
	errors.HandleError(err, "Unable to convert string value (%s) of env var (%s) to integer", true)

	return value
}

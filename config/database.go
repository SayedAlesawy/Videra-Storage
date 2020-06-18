package config

import (
	"sync"
)

// Databaseconfig Houses the configurations of the database
type Databaseconfig struct {
	User     string //Database user
	Password string //Database password
	Host     string //Database host
	Port     string //Database port
}

// databaseConfigOnce Used to garauntee thread safety for singleton instances
var databaseConfigOnce sync.Once

// nameNodeConfigInstance A singleton instance of the database object
var databaseConfigInstance *Databaseconfig

// DatabaseConfig A function to load database config
func (manager *ConfigurationManager) DatabaseConfig() *Databaseconfig {
	databaseConfigOnce.Do(func() {
		databaseConfig := Databaseconfig{
			User:     envString("DB_USER", "root"),
			Password: envString("DB_PASSWORD", "mysqlpassword"),
			Host:     envString("DB_HOST", "127.0.0.1"),
			Port:     envString("DB_PORT", "3306"),
		}

		databaseConfigInstance = &databaseConfig
	})

	return databaseConfigInstance
}

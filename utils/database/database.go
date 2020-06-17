package database

import (
	"fmt"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	_ "github.com/go-sql-driver/mysql" //mysql driver
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //gorm-mysql-dialect
)

// databaseOnce Used to garauntee thread safety for singleton instances
var databaseOnce sync.Once

// nameNodeConfigInstance A singleton instance of the database object
var databaseInstance *Database

// DBInstance A function to return a database instance
func DBInstance() *Database {
	databaseOnce.Do(func() {
		databaseConfig := config.ConfigurationManagerInstance("").DatabaseConfig()

		databaseObj := Database{
			User:     databaseConfig.User,
			Password: databaseConfig.Password,
			Host:     databaseConfig.Host,
			Port:     databaseConfig.Port,
			Name:     databaseConfig.Name,
		}

		dbHandler, err := gorm.Open("mysql", databaseObj.connectionString())
		if errors.IsError(err) {
			panic(err)
		}

		databaseObj.Connection = dbHandler

		databaseInstance = &databaseObj
	})

	return databaseInstance
}

// connectionString A function to return the database connection string
func (db *Database) connectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True",
		db.User, db.Password, db.Host, db.Port, db.Name)
}

package database

import "github.com/jinzhu/gorm"

// Database Houses database connection handler
type Database struct {
	User       string   //Database user
	Password   string   //Database password
	Host       string   //Database host
	Port       string   //Database port
	Name       string   //Database name
	Connection *gorm.DB //Connection to the data base
}

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func main() {
	logger := log.New(os.Stdout, "apiserver", log.LstdFlags)

	dbConfig, err := getDBConnConfig()
	if err != nil {
		logger.Printf("Error getting db connection information: %v", err)
		return
	}
	accountService, err := NewAccountService(dbConfig, logger)
	if err != nil {
		logger.Printf("Error connecting to account service: %v", err)
		return
	}
	defer accountService.Close()

	router := gin.Default()
	v1 := router.Group("v1")
	{
		v1.GET("/health", health)
		SetupAccountRouting(v1.Group("account"), accountService, logger)
	}
	router.Run()
}

func getDBConnConfig() (*DBConfig, error) {
	config := DBConfig{}
	var ok bool

	if config.Host, ok = os.LookupEnv("DB_HOST"); !ok {
		return nil, fmt.Errorf("Error: DB_HOST env variable not set")
	}

	if port, ok := os.LookupEnv("DB_PORT"); ok {
		var err error
		if config.Port, err = strconv.Atoi(port); err != nil {
			return nil, fmt.Errorf("Error: DB_PORT env variable is not a number %v", port)
		}
	} else {
		return nil, fmt.Errorf("Error: DB_PORT env variable not set")
	}

	if config.User, ok = os.LookupEnv("DB_USER"); !ok {
		return nil, fmt.Errorf("Error: DB_USER env variable not set")
	}
	if config.Password, ok = os.LookupEnv("DB_PASSWORD"); !ok {
		return nil, fmt.Errorf("Error: DB_PASSWORD env variable not set")
	}
	if config.Database, ok = os.LookupEnv("DB_DATABASE"); !ok {
		return nil, fmt.Errorf("Error: DB_DATABASE env variable not set")
	}

	return &config, nil
}

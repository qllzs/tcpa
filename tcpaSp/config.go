package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//GViperCfg  hss.json cfd handle
var GViperCfg *viper.Viper

func init() {
	GViperCfg = viper.New()
	GViperCfg.SetConfigFile("/opt/nkt/tcpaSp/config/tcpaSp.json")

	err := GViperCfg.ReadInConfig() // Find and read the config file
	if err != nil {                 // Handle errors reading the config file
		log.Fatalln("Fatal error config file: ", err)
	}
}

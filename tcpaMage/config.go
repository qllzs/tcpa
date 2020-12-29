package main

import (
	"fmt"

	"github.com/spf13/viper"
)

//GViperCfg  hss.json cfd handle
var GViperCfg *viper.Viper

func init() {
	GViperCfg = viper.New()
	GViperCfg.SetConfigFile("/opt/nkt/tcpaMage/config/tcpaMage.json")

	err := GViperCfg.ReadInConfig() // Find and read the config file
	if err != nil {                 // Handle errors reading the config file
		fmt.Println("Fatal error config file: ", err)
	}
}

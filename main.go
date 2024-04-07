package main

import (
	"flag"
	"log"
	"os"
	radius "radius-server/pkg/radius/server"
	"radius-server/pkg/usercenter"
	"time"
)

var userLoginCacheTime = flag.Duration("cache", time.Minute, "user login cache time")
var radiusSecret = flag.String("secret", "test123", "radius server secret")
var bypassPassword = flag.String("bypass", "", "radius bypass password")
var configPath = flag.String("configpath", "userconfig.yaml", "user config file")
var ListenAddr = flag.String("listenaddr", "0.0.0.0:1812", "radius server listen addr")

func main() {
	flag.Parse()
	_, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	radius.New(*radiusSecret, *bypassPassword, *userLoginCacheTime,
		usercenter.UserFromFileConfig(*configPath)).Serve(*ListenAddr)
}

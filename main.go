package main

import (
	"flag"
	"fmt"
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
var authAddr = flag.String("authaddr", "0.0.0.0:1812", "radius server listen addr")
var acctAddr = flag.String("acctaddr", "0.0.0.0:1813", "radius server listen addr")
var metricsAddr = flag.String("metricsaddr", "0.0.0.0:9000", "radius server listen addr")

func main() {
	flag.Parse()
	_, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if *bypassPassword != "" {
		fmt.Println("bypass password:", *bypassPassword)
	}
	fmt.Println("radius server secret:", *radiusSecret)
	rs := radius.New(*radiusSecret, *bypassPassword, *userLoginCacheTime,
		usercenter.UserFromFileConfig(*configPath))
	go rs.AcctServe(*acctAddr, *metricsAddr)
	rs.Serve(*authAddr)
}

package main

import (
	"context"
	"flag"
	"log"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

var server = flag.String("server", "localhost", "radius server addr")
var username = flag.String("u", "69789", "login user")
var password = flag.String("p", "111111", "login password")
var secret = flag.String("secret", "test123", "radius secret")

func main() {
	flag.Parse()
	packet := radius.New(radius.CodeAccessRequest, []byte(*secret))
	rfc2865.UserName_SetString(packet, *username)
	rfc2865.UserPassword_SetString(packet, *password)
	rfc2865.NASIdentifier_AddString(packet, "test")
	rfc2865.CallingStationID_AddString(packet, "localhost")
	rfc2869.ConnectInfo_AddString(packet, "radius-test-client")
	response, err := radius.Exchange(context.Background(), packet, *server+":1812")
	if err != nil {
		log.Fatal(err)
	}
	routes, _ := rfc2865.FramedRoute_GetStrings(response)
	log.Printf("Code:%v,   Attributes: %v,  routes %v", response.Code, response.Attributes, routes)
}

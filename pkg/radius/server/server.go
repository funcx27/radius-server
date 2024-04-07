package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func New(secret, bypassPassword string, cacheTime time.Duration, userService UserService) *RadiusServer {
	return &RadiusServer{
		secret:             secret,
		userLoginCacheTime: cacheTime,
		bypassPassword:     bypassPassword,
		userLoginCache:     map[string]string{},
		userService:        userService,
	}
}

func (rs *RadiusServer) Serve(ListenAddr string) {
	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(rs.authHandler()),
		SecretSource: radius.StaticSecretSource([]byte(rs.secret)),
	}
	if err := serve(&server, ListenAddr); err != nil {
		log.Fatal(err)
	}
}

func serve(s *radius.PacketServer, ListenAddr string) error {
	if s.Handler == nil {
		return errors.New("radius: nil Handler")
	}
	if s.SecretSource == nil {
		return errors.New("radius: nil SecretSource")
	}
	if s.Addr != "" {
		ListenAddr = s.Addr
	}
	network := "udp4"
	if s.Network != "" {
		network = s.Network
	}
	pc, err := net.ListenPacket(network, ListenAddr)
	if err != nil {
		return err
	}
	fmt.Println("Starting server on", ListenAddr)
	defer pc.Close()
	return s.Serve(pc)
}

func (rs *RadiusServer) authHandler() func(w radius.ResponseWriter, r *radius.Request) {
	return func(w radius.ResponseWriter, r *radius.Request) {
		username := rfc2865.UserName_GetString(r.Packet)
		password := rfc2865.UserPassword_GetString(r.Packet)
		nasIdentifier := rfc2865.NASIdentifier_GetString(r.Packet)
		clientAddr := rfc2865.CallingStationID_GetString(r.Packet)
		clientSoftwareVersion := rfc2869.ConnectInfo_GetString(r.Packet)
		code, routes, group, err := rs.responseCode(username, password, nasIdentifier)
		rs.userService.LoginInfoHandler(username, password, nasIdentifier, clientAddr, clientSoftwareVersion, code.String(), err)
		responsePacket := r.Response(code)
		if group != "" {
			if err := rfc2865.Class_AddString(responsePacket, "OU="+group); err != nil {
				fmt.Printf("rfc2865 Class add error: %v\n", err)
			}
		}
		for _, route := range routes {
			if err := rfc2865.FramedRoute_AddString(responsePacket, route); err != nil {
				fmt.Printf("rfc2865 Frame-Route add error: %v\n", err)
			}
		}
		w.Write(responsePacket)
	}
}

func (rs *RadiusServer) responseCode(username, password, nasIdentifier string) (radius.Code, []string, string, error) {
	routes := rs.userService.UserRoutesQuery(username, nasIdentifier)
	group := rs.userService.UserGroupQuery(username, nasIdentifier)
	if len(routes) == 0 && group == "" {
		return radius.CodeAccessReject, routes, group, errors.New("user routes or group not found")
	}
	if len(password) < 6 {
		return radius.CodeAccessReject, routes, group, errors.New("password lens less than 6")
	}
	if rs.bypassPassword != "" && password == rs.bypassPassword {
		// log.Printf("username %s login success with bypass password", username)
		return radius.CodeAccessAccept, routes, group, nil
	}
	if password == rs.userLoginCache[username] {
		// log.Printf("username %s login success with cache otp %s", username, password)
		return radius.CodeAccessAccept, routes, group, nil
	}
	AuthenticationError := rs.userService.Authentication(username, password)
	if len(password) == 6 && AuthenticationError == nil {
		// log.Printf("username %s login success with otp %s", username, password)
		go func() {
			rs.userLoginCache[username] = password
			timer := time.NewTimer(rs.userLoginCacheTime)
			<-timer.C
			delete(rs.userLoginCache, username)
		}()
		return radius.CodeAccessAccept, routes, group, nil
	}
	return radius.CodeAccessReject, routes, group, AuthenticationError
}

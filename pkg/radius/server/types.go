package server

import (
	"time"
)

type RadiusServer struct {
	secret             string
	bypassPassword     string
	userLoginCache     map[string]string
	userLoginCacheTime time.Duration
	userService        UserService
}

type UserService interface {
	UserRoutesQuery(user, nasIdentifier string) (routes []string) // ocserv auth config groupconfig=true
	UserGroupQuery(username, nasIdentifier string) string         // ocserv auth config groupconfig=false
	Authentication(username, password string) error
	LoginInfoHandler(username, password, nasIdentifier, clientAddr, clientSoftwareVersion, radiusCode string, err error)
}

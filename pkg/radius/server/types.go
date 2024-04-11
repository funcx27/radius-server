package server

import (
	"net"
	"radius-server/pkg/exporter"
	"time"

	"layeh.com/radius/rfc2866"
)

type RadiusServer struct {
	secret             string
	bypassPassword     string
	userLoginCache     map[string]string
	userLoginCacheTime time.Duration
	userService        UserService
	*exporter.Exporter
}

type UserService interface {
	UserRoutesQuery(user, nasIdentifier string) (routes []string) // ocserv auth config groupconfig=true
	UserGroupQuery(username, nasIdentifier string) string         // ocserv auth config groupconfig=false
	Authentication(username, password string) error
	LoginInfoHandler(username, password, nasIdentifier, clientAddr, clientSoftwareVersion, radiusCode string, err error)
}

type AccountingSession struct {
	SessionId       string
	Username        string
	NasIdentifier   string
	Status          rfc2866.AcctStatusType
	RemoteIpAddress string
	FramedIPAddress net.IP
	SessionTime     uint32
	InputBytes      uint64
	OutputBytes     uint64
	TerminateCause  rfc2866.AcctTerminateCause
}

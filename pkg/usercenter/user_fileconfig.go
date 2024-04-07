package usercenter

import (
	"log"
	"os"
	radius "radius-server/pkg/radius/server"

	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	configFilePath string
	Routes         map[string][]string `yaml:"routes"`
	UserRoutes     map[string][]string `yaml:"users"`
	GroupUsers     map[string][]string `yaml:"groups"`
}

func UserFromFileConfig(file string) radius.UserService {
	return &fileConfig{
		configFilePath: file,
	}
}

func (cfg *fileConfig) Load() error {
	b, err := os.ReadFile(cfg.configFilePath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, &cfg)

}

func (cfg *fileConfig) UserRoutesQuery(username, nasIdentifier string) (routes []string) {
	cfg.Load()
	for _, group := range cfg.UserRoutes[username+"@"+nasIdentifier] {
		for _, route := range cfg.Routes[group] {
			routes = append(routes, route)
		}
	}
	return routes
}

func (cfg *fileConfig) UserGroupQuery(username, nasIdentifier string) string {
	cfg.Load()
	for group, users := range cfg.GroupUsers {
		for _, user := range users {
			if username+"@"+nasIdentifier == user {
				return group
			}
		}
	}
	return ""
}

func (cfg *fileConfig) Authentication(username, password string) error {
	return tianUserOtpLogin(username, password)

}

func (cfg *fileConfig) LoginInfoHandler(username, password, nasIdentifier, clientAddr, clientSoftwareVersion, radiusCode string, err error) {
	if err == nil {
		log.Printf("%s@%s, vpnclient %s(%s): %s", username, nasIdentifier, clientAddr, clientSoftwareVersion, radiusCode)
	} else {
		log.Printf("%s@%s, vpnclient %s(%s): %s, Cause by: %s", username, nasIdentifier, clientAddr, clientSoftwareVersion, radiusCode, err.Error())

	}
}

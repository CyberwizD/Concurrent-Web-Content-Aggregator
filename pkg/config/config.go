package config

import "github.com/CyberwizD/RESTful-API-with-JWT-auth-and-RBAC/config"

type Config struct {
	App
	sources
}

func LoadConfig(cg string, sc string) *config.Config {

}

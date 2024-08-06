package auth

import config "github.com/whitewhale1075/urmy_rest_config"

type AuthHandler struct {
	ConfigHandler *config.Handler
}

type InsertUserProfile struct {
	DeviceOS string `json:"deviceos"`
	LoginID  string `json:"loginId"`
	Useruid  string `json:"userUid"`
	Name     string `json:"name"`
}

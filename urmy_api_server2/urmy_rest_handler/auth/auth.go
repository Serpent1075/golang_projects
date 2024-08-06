package auth

import (
	"log"

	config "github.com/whitewhale1075/urmy_rest_config"
)

type AuthInterface interface {
	RegisterUser(signupuser *InsertUserProfile)
}

func (ah *AuthHandler) RegisterUser(signupuser *InsertUserProfile) {
	_, adduserinfoerr := ah.ConfigHandler.ReadDB.Exec("INSERT INTO urmyuserinfo (loginuuid, deviceos, email, name) VALUES ($1, $2, $3, $4)",
		signupuser.Useruid, signupuser.DeviceOS, signupuser.LoginID, signupuser.Name)
	if adduserinfoerr != nil {
		log.Println(adduserinfoerr)
	}

}

func NewAuthHandler(path string) *AuthHandler {
	return newAuthHandler(path)
}

func newAuthHandler(path string) *AuthHandler {
	handler := config.Init()
	log.Println(path)
	return &AuthHandler{
		ConfigHandler: handler,
	}
}

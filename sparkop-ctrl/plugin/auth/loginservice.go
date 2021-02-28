package auth

import (
	log "github.com/sirupsen/logrus"
	muser "github.com/spark-sql-on-k8s/pkg/user"
)

type LoginCredentials struct {
	Username    string `form:"username" json:"username"`
	Password    string `form:"password" json:"password"`
}

type LoginService interface {
	LoginUser(username string, password string) bool
}

type loginInformation struct {
	username    string
	password    string
}

func StaticLoginService() LoginService {
	return &loginInformation{
		username:    "default",
		password:    "",
	}
}

// TODO: 通过 database 授权
func BuildLoginService(conf map[string]string) LoginService {
	return &loginInformation{}
}

func (info *loginInformation) LoginUser(username string, password string) bool {	
	err, user := muser.ValidateUserPassword(username, password)

	log.Infof("current login : username: %s, pw: %s; result: err=%v, user=%v", username, password, err, user)

	if err != nil {
		return false
	}

	return true
}

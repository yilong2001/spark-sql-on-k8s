package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//login contorller interface
type LoginController interface {
	Login(ctx *gin.Context) (string, error)
}

type loginController struct {
	loginService LoginService
	jWtService   JWTService
}

var loginCtrl LoginController

func BuildLoginController(loginService LoginService,
	jWtService JWTService) LoginController {
	if (loginCtrl == nil) {
		loginCtrl = &loginController{
			loginService: loginService,
			jWtService:   jWtService,
		}
	}

	return loginCtrl
}

func (controller *loginController) Login(ctx *gin.Context) (string, error) {
	var credential LoginCredentials
	err := ctx.ShouldBind(&credential)
	if err != nil {
		log.Errorf("error: bind login credential : %v", err)
		return "", err
	}

	isUserAuthenticated := controller.loginService.LoginUser(credential.Username, 
		credential.Password)
	if isUserAuthenticated {
		return controller.jWtService.GenerateToken(credential.Username, true)
	}

	return "", fmt.Errorf("error: user auth failed : %s!", credential.Username)
}

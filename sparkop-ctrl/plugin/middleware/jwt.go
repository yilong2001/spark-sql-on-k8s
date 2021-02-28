package middleware

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	
	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	myauth "github.com/spark-sql-on-k8s/sparkop-ctrl/plugin/auth"
)

func AuthorizeJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) <= len(BEARER_SCHEMA) {
			log.Error("Authorization header is empty!")
			sacommon.NewGinFaliResponse(http.StatusUnauthorized, 
				"Authorization header is empty!","", c)
			c.Abort()
			return
		}

		tokenString := authHeader[len(BEARER_SCHEMA):]
		token, err := myauth.BuildJWTService().ValidateToken(tokenString)
		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			log.Info(claims)
			c.Set("X-User", claims["username"].(string))
			//c.Writer.Header().Set("X-User", claims["username"].(string))
		} else {
			log.Error(err)
			sacommon.NewGinFaliResponse(http.StatusUnauthorized, 
				fmt.Sprintf("%v",err),"", c)
			c.Abort()
		}
	}
}

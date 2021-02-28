package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)


//jwt service
type JWTService interface {
	GenerateToken(username string, isUser bool) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type authCustomClaims struct {
	Username string `json:"username"`
	Isok   bool   `json:"isok"`
	jwt.StandardClaims
}

type jwtServices struct {
	secretKey string
	issure    string
}

//auth-jwt
func BuildJWTService() JWTService {
	return &jwtServices{
		secretKey: getSecretKey(),
		issure:    "sparkapp",
	}
}

func getSecretKey() string {
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "abcdefghigklmnopqrstuvwxyz"
	}
	return secret
}

func (service *jwtServices) GenerateToken(username string, isUser bool) (string, error) {
	claims := &authCustomClaims{
		username,
		isUser,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			Issuer:    service.issure,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	return token.SignedString([]byte(service.secretKey))
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("Invalid token", token.Header["alg"])
		}
		
		return []byte(service.secretKey), nil
	})

}

package dutchman

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
)

var (
	ErrInvalid = errors.New("invalid token")
)

type Claims struct {
	jwt.StandardClaims
	ID    string `json:"id"`
	Email string `json:"username"`
}

func GenerateToken(user *models.User) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
	})

	str, err := token.SignedString([]byte("abracadabra"))
	if err != nil {
		return "", err
	}

	return str, nil
}

func ParseToken(accessToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}

		return []byte("abracadabra"), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalid
}

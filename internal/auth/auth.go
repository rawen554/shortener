package auth

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const TokenExp = time.Hour * 3
const SecretKey = "b4952c3809196592c026529df00774e46bfb5be0"
const CookieName = "jwt-token"
const UserIDKey = "userID"

var ErrTokenNotValid = errors.New("token is not valid")
var ErrNoUserInToken = errors.New("no user data in token")

func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: uuid.New().String(),
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		if !token.Valid {
			return "", ErrTokenNotValid
		} else {
			return "", errors.New("parsing error")
		}
	}

	if claims.UserID == "" {
		return "", ErrNoUserInToken
	}

	return claims.UserID, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(CookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				token, err := BuildJWTString()
				if err != nil {
					log.Printf("Error building JWT string: %v", err)
					c.Writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				c.SetCookie(CookieName, token, 3600*24*30, "", "", false, true)
				cookie = token
			} else {
				log.Printf("Error reading cookie[%v]: %v", CookieName, err)
				c.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		userID, err := GetUserID(cookie)
		if err != nil {
			if errors.Is(err, ErrNoUserInToken) {
				c.Writer.WriteHeader(http.StatusUnauthorized)
				return
			}
			if errors.Is(err, ErrTokenNotValid) {
				token, err := BuildJWTString()
				if err != nil {
					log.Printf("Error building JWT string: %v", err)
					c.Writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				userID, err = GetUserID(token)
				if err != nil {
					log.Printf("Revalidate error user id from renewed token: %v", err)
					c.Writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				c.SetCookie(CookieName, token, 3600*24*30, "", "", false, true)
			}
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

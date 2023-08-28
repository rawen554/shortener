// Модуль аутентификации клиентских запросов.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	tokenExp   = time.Hour * 3
	maxAge     = 3600 * 24 * 30
	cookieName = "jwt-token"
	UserIDKey  = "userID"
)

var ErrTokenNotValid = errors.New("token is not valid")
var ErrNoUserInToken = errors.New("no user data in token")
var ErrBuildJWTString = errors.New("error building JWT string")

func BuildJWTString(secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: uuid.New().String(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error creating signed JWT: %w", err)
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(tokenString string, secret string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
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

func AuthMiddleware(secret string, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				token, err := BuildJWTString(secret)
				if err != nil {
					logger.Error(ErrBuildJWTString, err)
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}
				c.SetCookie(cookieName, token, maxAge, "", "", false, true)
				cookie = token
			} else {
				logger.Error("Error reading cookie[%v]: %v", cookieName, err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		userID, err := GetUserID(cookie, secret)
		if err != nil {
			if errors.Is(err, ErrNoUserInToken) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if errors.Is(err, ErrTokenNotValid) {
				token, err := BuildJWTString(secret)
				if err != nil {
					logger.Error(ErrBuildJWTString, err)
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}
				userID, err = GetUserID(token, secret)
				if err != nil {
					logger.Error("Revalidate error user id from renewed token: %v", err)
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}
				c.SetCookie(cookieName, token, maxAge, "", "", false, true)
			}
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

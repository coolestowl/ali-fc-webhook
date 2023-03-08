package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwtSecret       = []byte("default-key-github.com/golang-jwt/jwt/v4")
	errUnauthorized = errors.New("Unauthorized")
)

func InitJwtSecret(key []byte) {
	jwtSecret = key
}

func JWTAuth(ctx *gin.Context) {
	f := func() error {
		token := ctx.Query("token")
		if len(token) == 0 {
			return errUnauthorized
		}

		tk, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("mismatched signing method")
			}
			return jwtSecret, nil
		})
		if err != nil {
			return err
		}

		claims, ok := tk.Claims.(jwt.MapClaims)
		if !ok || !tk.Valid {
			return errUnauthorized
		}

		if claims["dat"] != "ali-fc-webhook" {
			return errUnauthorized
		}
		return nil
	}

	if err := f(); err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	ctx.Next()
}

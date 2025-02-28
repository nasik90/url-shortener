package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const tokenExp = time.Hour * 3
const secretKey = "supersecretkey"

// BuildJWTString создаёт токен и возвращает его в виде строки.
func buildJWTString() (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// Для данной задачи достаточно UserID формировать на основе момента времени
		UserID: strconv.Itoa(int(time.Now().UnixMilli())),
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// cookies := r.Header.Get("Authorization")
		// fmt.Println(cookies)
		const (
			cookieName = "auth"
		)
		var userID = ""
		authCookieIn, err := req.Cookie(cookieName)
		if err == nil {
			userID, err = getUserID(authCookieIn.Value)
		}
		if err != nil {
			JWT, _ := buildJWTString()
			var authCookieOut http.Cookie
			authCookieOut.Name = cookieName
			authCookieOut.Value = JWT
			http.SetCookie(res, &authCookieOut)
			userID, err = getUserID(JWT)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if userID == "" {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(req.Context(), settings.UserIDContextKey, userID)
		req = req.WithContext(ctx)
		h.ServeHTTP(res, req)
	}
}

func getUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		// fmt.Println("Token is not valid")
		return "", fmt.Errorf("token is not valid")
	}

	// fmt.Println("Token os valid")
	return claims.UserID, nil
}

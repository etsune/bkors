package main

import (
	"fmt"

	"github.com/etsune/bkors/config"
	"github.com/etsune/bkors/services"
	"github.com/labstack/echo/v4"
)

func SetRequestUser(s *services.UserService, config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// валидация токена
			// дешифровка токена
			jwt, err := ctx.Cookie("access_token")
			if err != nil {
				fmt.Printf("1%v/n", err)
				return next(ctx)
			}
			userId, err := services.ParseToken(jwt.Value, config.AccessTokenPrivateKey)
			if err != nil {
				fmt.Printf("2%v/n", err)
				return next(ctx)
			}

			// получение пользователя
			user, err := s.GetUser(userId)
			if err != nil {
				fmt.Printf("3%v/n", err)
				return next(ctx)
			}

			// добавление пользователя в контекст
			ctx.Set("userdata", user)

			return next(ctx)
		}
	}
}

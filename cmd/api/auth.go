package main

import (
	"bytes"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

var jwtSecret []byte

func login(c echo.Context) error {
	req := struct {
		MailAddress string `json:"mailAddress"`
	}{}
	err := c.Bind(&req)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	claims := &jwt.StandardClaims{
		ExpiresAt: 15000,
		Issuer:    "statuspage",
		Audience:  "statuspage",
		Subject:   req.MailAddress,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	httpReq, err := http.NewRequest("POST", "http://localhost:3002", bytes.NewBufferString(tokenStr))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	_, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	c.Logger().Error(tokenStr, err)
	return echo.NewHTTPError(200)
}

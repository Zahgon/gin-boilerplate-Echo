package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// AuthController ...
type AuthController struct{}

var authModel = new(models.AuthModel)

// Refresh Token godoc
// @Summary Refresh Token example
// @Schemes
// @Description Refresh Token example
// @Tags Auth
// @Accept json
// @Produce json
// @Param auth body forms.Token true "Auth"
// @Success 	 200  {object}  models.AuthResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /token/refresh [POST]
func (ctl AuthController) Refresh(c echo.Context) error {
	var tokenForm forms.Token

	if c.Bind(&tokenForm) != nil || c.Validate(&tokenForm) != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": "Invalid form", "form": tokenForm})
	}

	//verify the token
	token, err := jwt.Parse(tokenForm.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid authorization, please login again"})
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid authorization, please login again"})
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid authorization, please login again"})
		}
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid authorization, please login again"})
		}
		//Delete the previous Refresh Token
		deleted, delErr := authModel.DeleteAuth(refreshUUID)
		if delErr != nil || deleted == 0 { //if any goes wrong
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid authorization, please login again"})
		}

		//Create new pairs of refresh and access tokens
		ts, createErr := authModel.CreateToken(userID)
		if createErr != nil {
			return c.JSON(http.StatusForbidden, echo.Map{"message": "Invalid authorization, please login again"})
		}
		//save the tokens metadata to redis
		saveErr := authModel.CreateAuth(userID, ts)
		if saveErr != nil {
			return c.JSON(http.StatusForbidden, echo.Map{"message": "Invalid authorization, please login again"})
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		return c.JSON(http.StatusOK, tokens)
	}

	return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid authorization, please login again"})
}

package controllers

import (
	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"

	"net/http"

	"github.com/labstack/echo/v4"
)

// UserController ...
type UserController struct{}

var userModel = new(models.UserModel)

// getUserID ...
func getUserID(c echo.Context) (userID int64) {
	return c.Get("userID").(int64)
}

// Login User godoc
// @Summary Login User example
// @Schemes
// @Description Login User example
// @Tags User
// @Accept json
// @Produce json
// @Param article body forms.LoginForm true "User"
// @Success 	 200  {object}  models.UserLoginResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/login [post]
func (ctrl UserController) Login(c echo.Context) error {
	var loginForm forms.LoginForm

	if validationErr := c.Bind(&loginForm); validationErr != nil {
		message := forms.Translate(validationErr, forms.LoginMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	if validationErr := c.Validate(&loginForm); validationErr != nil {
		message := forms.Translate(validationErr, forms.LoginMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	user, token, err := userModel.Login(loginForm)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": "Invalid login details"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged in", "user": user, "token": token})
}

// Register User godoc
// @Summary Register User example
// @Schemes
// @Description Register User example
// @Tags User
// @Accept json
// @Produce json
// @Param article body forms.RegisterForm true "User"
// @Success 	 200  {object}  models.UserLoginResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/register [post]
func (ctrl UserController) Register(c echo.Context) error {
	var registerForm forms.RegisterForm

	if validationErr := c.Bind(&registerForm); validationErr != nil {
		message := forms.Translate(validationErr, forms.RegisterMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	if validationErr := c.Validate(&registerForm); validationErr != nil {
		message := forms.Translate(validationErr, forms.RegisterMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	user, err := userModel.Register(registerForm)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully registered", "user": user})
}

// Logout User godoc
// @Summary Logout User example
// @Schemes
// @Description Logout User example
// @Tags User
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/logout [GET]
func (ctrl UserController) Logout(c echo.Context) error {
	au, err := authModel.ExtractTokenMetadata(c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "User not logged in"})
	}

	deleted, delErr := authModel.DeleteAuth(au.AccessUUID)
	if delErr != nil || deleted == 0 {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid request"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
}

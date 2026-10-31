package controllers

import (
	"strconv"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"

	"net/http"

	"github.com/labstack/echo/v4"
)

// ArticleController ...
type ArticleController struct{}

var articleModel = new(models.ArticleModel)

// Create Article godoc
// @Summary Create Article example
// @Schemes
// @Description Create Article example
// @Tags Article
// @Accept json
// @Produce json
// @Param article body forms.CreateArticleForm true "Article"
// @Success 	 200  {object}  forms.ArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article [post]
func (ctrl ArticleController) Create(c echo.Context) error {
	userID := getUserID(c)

	var form forms.CreateArticleForm

	if validationErr := c.Bind(&form); validationErr != nil {
		message := forms.Translate(validationErr, forms.ArticleMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	if validationErr := c.Validate(&form); validationErr != nil {
		message := forms.Translate(validationErr, forms.ArticleMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	id, err := articleModel.Create(userID, form)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": "Article could not be created"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Article created", "id": id})
}

// Get All Articles godoc
// @Summary Get All Articles example
// @Schemes
// @Description Get All Articles example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.AllArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /articles [GET]
func (ctrl ArticleController) All(c echo.Context) error {
	userID := getUserID(c)

	results, err := articleModel.All(userID)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"Message": "Could not get articles"})
	}

	return c.JSON(http.StatusOK, echo.Map{"results": results})
}

// Get One Article godoc
// @Summary Get One Article example
// @Schemes
// @Description One All Article example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.OneArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article/{id} [GET]
func (ctrl ArticleController) One(c echo.Context) error {
	userID := getUserID(c)

	id := c.Param("id")

	getID, err := strconv.ParseInt(id, 10, 64)
	if getID == 0 || err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"Message": "Invalid parameter"})
	}

	data, err := articleModel.One(userID, getID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"Message": "Article not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// Update Article godoc
// @Summary Update Article example
// @Schemes
// @Description Update Article example
// @Tags Article
// @Accept json
// @Produce json
// @Param article body forms.CreateArticleForm true "Article"
// @Success 	 200  {object}  models.ArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article/{id} [PUT]
func (ctrl ArticleController) Update(c echo.Context) error {
	userID := getUserID(c)

	id := c.Param("id")

	getID, err := strconv.ParseInt(id, 10, 64)
	if getID == 0 || err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"Message": "Invalid parameter"})
	}

	var form forms.CreateArticleForm

	if validationErr := c.Bind(&form); validationErr != nil {
		message := forms.Translate(validationErr, forms.ArticleMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	if validationErr := c.Validate(&form); validationErr != nil {
		message := forms.Translate(validationErr, forms.ArticleMessages)
		return c.JSON(http.StatusNotAcceptable, echo.Map{"message": message})
	}

	err = articleModel.Update(userID, getID, form)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"Message": "Article could not be updated"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Article updated"})
}

// Delete Article godoc
// @Summary Delete Article example
// @Schemes
// @Description Delete Article example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.OneArticleResponse
// @Success 	 404  {object}  forms.ArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article/{id} [DELETE]
func (ctrl ArticleController) Delete(c echo.Context) error {
	userID := getUserID(c)

	id := c.Param("id")

	getID, err := strconv.ParseInt(id, 10, 64)
	if getID == 0 || err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"Message": "Invalid parameter"})
	}

	err = articleModel.Delete(userID, getID)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, echo.Map{"Message": "Article could not be deleted"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Article deleted"})
}

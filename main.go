package main

import (
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/Massad/gin-boilerplate/controllers"
	"github.com/Massad/gin-boilerplate/db"
	_ "github.com/Massad/gin-boilerplate/docs"
	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/invoice"
	"github.com/Massad/gin-boilerplate/middleware"
	"github.com/Massad/gin-boilerplate/renderer"
	"github.com/Massad/gin-boilerplate/router"
	"github.com/Massad/gin-boilerplate/swagger"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

// @title           Golang Gin Boilerplate
// @version         3.0
// @description     A RESTful API boilerplate with Gin Framework, PostgreSQL, Redis and JWT authentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT License
// @license.url   https://github.com/Massad/gin-boilerplate/blob/master/LICENSE

// @host      localhost:9000
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error: failed to load the env file")
	}

	r := echo.New()
	r.HideBanner = true
	r.HidePort = true
	r.IPExtractor = echo.ExtractIPDirect()

	r.Binder = new(forms.JSONBinder)
	r.Validator = new(forms.DefaultValidator)
	r.JSONSerializer = new(renderer.JSONSerializer)
	r.Renderer = renderer.New("./public/html/*")

	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recover())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Gzip())

	db.Init()
	db.InitRedis(1)

	v1 := r.Group("/v1")
	{
		/*** START USER ***/
		user := new(controllers.UserController)

		v1.POST("/user/login", user.Login)
		v1.POST("/user/register", user.Register)
		v1.GET("/user/logout", user.Logout, middleware.TokenAuth())

		/*** START AUTH ***/
		auth := new(controllers.AuthController)

		v1.POST("/token/refresh", auth.Refresh)

		/*** START Article ***/
		article := new(controllers.ArticleController)

		v1.POST("/article", article.Create, middleware.TokenAuth())
		v1.GET("/articles", article.All, middleware.TokenAuth())
		v1.GET("/article/:id", article.One, middleware.TokenAuth())
		v1.PUT("/article/:id", article.Update, middleware.TokenAuth())
		v1.DELETE("/article/:id", article.Delete, middleware.TokenAuth())

		/*** START Invoice ***/
		inv := new(invoice.InvoiceController)

		v1.GET("/invoice", inv.Preview)
		v1.GET("/invoice/download", inv.Download)
	}

	r.GET("/swagger/*", swagger.Handler())

	static := middleware.StaticFiles("/public", "./public")
	r.GET("/public/*", static)
	r.HEAD("/public/*", static)

	r.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", echo.Map{
			"ginBoilerplateVersion": "v3.0",
			"goVersion":             runtime.Version(),
		})
	})

	table := routeTable(r)
	r.Pre(middleware.DecodedPathRouting())
	r.Pre(middleware.TrailingSlashRedirect(table))
	r.Pre(middleware.DeferredStatus())
	r.Use(middleware.RouteGuard(table))

	reportError := r.HTTPErrorHandler
	r.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		//Unmatched routes and unmatched methods both fall through to the 404 page
		if httpErr, ok := err.(*echo.HTTPError); ok &&
			(httpErr.Code == http.StatusNotFound || httpErr.Code == http.StatusMethodNotAllowed) {
			c.Response().Header().Del(echo.HeaderAllow)
			if renderErr := c.Render(http.StatusNotFound, "404.html", echo.Map{}); renderErr == nil {
				return
			}
		}
		reportError(err, c)
	}

	port := os.Getenv("PORT")

	log.Printf("\n\n PORT: %s \n ENV: %s \n SSL: %s \n Version: %s \n\n", port, os.Getenv("ENV"), os.Getenv("SSL"), os.Getenv("API_VERSION"))

	if os.Getenv("SSL") == "TRUE" {
		SSLKeys := &struct {
			CERT string
			KEY  string
		}{
			CERT: "./cert/myCA.cer",
			KEY:  "./cert/myCA.key",
		}

		r.Logger.Fatal(r.StartTLS(":"+port, SSLKeys.CERT, SSLKeys.KEY))
	} else {
		r.Logger.Fatal(r.Start(":" + port))
	}
}

func routeTable(r *echo.Echo) *router.Table {
	routes := make(map[string][]string)
	for _, route := range r.Routes() {
		routes[route.Method] = append(routes[route.Method], route.Path)
	}
	return router.New(routes)
}

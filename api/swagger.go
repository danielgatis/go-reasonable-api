package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	swaggerFiles "github.com/swaggo/files/v2"
	"github.com/swaggo/swag"
)

var swaggerFileServer = http.FileServer(http.FS(swaggerFiles.FS))

// swaggerHandler returns an echo.HandlerFunc that serves the Swagger UI.
// It bridges the net/http handler from swaggo/files to echo v5.
func swaggerHandler(c *echo.Context) error {
	req := c.Request()

	// Strip /swagger prefix so the file server can find files
	path := strings.TrimPrefix(req.URL.Path, "/swagger")
	if path == "" || path == "/" {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	}

	if path == "/doc.json" {
		doc, err := swag.ReadDoc()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
		return c.String(http.StatusOK, doc)
	}

	req.URL.Path = path
	swaggerFileServer.ServeHTTP(c.Response(), req)
	return nil
}

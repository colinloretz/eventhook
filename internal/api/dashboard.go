package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeDashboard registers the embedded React SPA at /dashboard and /dashboard/*.
// assets must have the built files under a "dashboard" subdirectory.
func (s *Server) ServeDashboard(assets embed.FS) {
	sub, err := fs.Sub(assets, "dashboard")
	if err != nil {
		panic("dashboard assets not found: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	handler := func(c *gin.Context) {
		// Strip /dashboard prefix so the file server maps into the dist root
		p := strings.TrimPrefix(c.Request.URL.Path, "/dashboard")
		if p == "" {
			p = "/"
		}
		c.Request.URL.Path = p
		fileServer.ServeHTTP(c.Writer, c.Request)
	}

	s.router.GET("/dashboard", handler)
	s.router.GET("/dashboard/*any", handler)
}

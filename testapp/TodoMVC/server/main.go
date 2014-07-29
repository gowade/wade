package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pilu/fresh/runner/runnerutils"
)

var (
	g *Environment = environment()
)

type Environment struct {
	devMode bool
}

func environment() *Environment {
	return &Environment{
		devMode: true,
	}
}

func (g *Environment) IsDevMode() bool {
	return g.devMode
}

func errPanic(err error, message string) {
	if err != nil {
		if message == "" {
			message = err.Error()
		}
		log.Printf(message)
		if g.IsDevMode() {
			panic(err.Error())
		}
	}
}

func checkErr(err error) {
	errPanic(err, "")
}

func runnerMiddleware(c *gin.Context) {
	if runnerutils.HasErrors() {
		runnerutils.RenderError(c.Writer)
		c.Abort(500)
	}
}

func main() {
	r := gin.Default()

	// IMPORTANT PART STARTS HERE(=
	//
	if g.IsDevMode() {
		r.Use(runnerMiddleware)
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			r.ServeFiles("/gopath/*filepath", http.Dir(gopath))
		}

		goroot := os.Getenv("GOROOT")
		if goroot != "" {
			r.ServeFiles("/goroot/*filepath", http.Dir(goroot))
		}
	}

	// This serves static files in the "public" directory
	r.ServeFiles("/public/*filepath", http.Dir("../public"))

	// Subpaths of /todo/ are client urls, should NOT be protected
	// Just serve the index.html for every subpaths actually, nothing else
	web := r.Group("/todo/", func(c *gin.Context) {
		f, err := os.Open("../public/index.html")
		checkErr(err)
		conts, err := ioutil.ReadAll(f)
		checkErr(err)
		c.Data(200, "text/html;charset=utf-8", conts)
	})
	web.GET("*path", func(c *gin.Context) {})

	// Redirect the home page to /todo/
	r.GET("/", func(c *gin.Context) {
		http.Redirect(c.Writer, c.Request, "/todo/", http.StatusFound)
	})

	//
	// =)IMPORTANT PART ENDS HERE

	r.Run(":3000")
}

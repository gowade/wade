package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jmcvetta/randutil"
	"github.com/phaikawl/wade/testapp/ez/model"
	"github.com/pilu/fresh/runner/runnerutils"
)

const (
	mySigningKey = "n0t9r34t6cz9r34tn0t1na9r34tw4y"
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

func makeRandomUserToken() (username string, tokenString string) {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	username, err := randutil.AlphaStringRange(5, 10)
	errPanic(err, "Cannot random string, wtf?")
	token.Claims["username"] = username
	token.Claims["secret"] = time.Now().Add(time.Hour * 72).Unix()
	tokenString, err = token.SignedString([]byte(mySigningKey))
	errPanic(err, "Cannot sign string, wtf?")
	return
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

	// Subpaths of /api/ provides the server API
	// they should be protected by authorization
	api := r.Group("/api/", func(c *gin.Context) {
		token := c.Request.Header.Get("AuthToken")
		if token != "" {
			return
		}

		c.Fail(http.StatusUnauthorized, fmt.Errorf("You're not allowed to do this, sorry."))
	})

	// The api to register a user
	// Used in the pg-user-register page
	api.POST("/user/register", func(c *gin.Context) {
		rawdata, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err.Error())
		}
		udata := &model.User{}
		if err = json.Unmarshal(rawdata, udata); err != nil {
			panic(err.Error())
		}

		failed := model.UsernamePasswordValidator().Validate(udata).HasErrors()
		c.JSON(200, !failed)
	})

	// Subpaths of /web/ are client urls, should NOT be protected
	// Just serve the index.html for every subpaths actually, nothing else
	web := r.Group("/web/", func(c *gin.Context) {
		f, err := os.Open("../public/index.html")
		checkErr(err)
		conts, err := ioutil.ReadAll(f)
		checkErr(err)
		c.Data(200, "text/html;charset=utf-8", conts)
	})
	web.GET("*path", func(c *gin.Context) {})

	// Redirect the home page to /web/
	r.GET("/", func(c *gin.Context) {
		http.Redirect(c.Writer, c.Request, "/web/", http.StatusFound)
	})

	//
	// =)IMPORTANT PART ENDS HERE

	r.GET("/auth", func(c *gin.Context) {
		username, token := makeRandomUserToken()
		user := &model.User{
			Username: username,
			Token:    token,
			Role:     model.RoleUser,
		}
		c.JSON(200, map[string]interface{}{
			"username": user.Username,
			"token":    user.Token,
		})
	})

	r.Run(":3000")
}

package main

import (
	wd "github.com/phaikawl/wade"
	"github.com/phaikawl/wade/services/http"
	"github.com/phaikawl/wade/services/pdata"
	"github.com/phaikawl/wade/testapp/ez/model"
	"github.com/phaikawl/wade/utils"
)

type UserInfo struct {
	Name string
	Age  int
}

type AuthedStat struct {
	AuthGened bool
}

type UsernamePassword struct {
	Username string
	Password string
}

// RegUser is a form model
// It embeds Wade's utils.Validated, which holds an Errors field that stores
// validation errors.
// Every form model should have a separate Data field for the data fields that are
// to be validated. Here we have the Data as UsernamePassword.
// Methods of a model can be accessed normally in HTML binding code, just like
// member variables
type RegUser struct {
	utils.Validated
	Data UsernamePassword
}

// This method is used for the bind-on-click of the Reset button
// to reset the form
func (r *RegUser) Reset() {
	r.Data.Password = ""
	r.Data.Username = ""
}

// This method is used for the bind-on-click of the Submit button
// to validate and send the form
func (r *RegUser) Submit() {
	// Wade provides a convenient way to send forms.
	// Just provide the target url, the Data, the model and a validator.
	// The ProcessForm function automatically does the validation and populates
	// r.Validated.Errors with the validation errors.
	// It returns the channel with the server's response, but we don't care for now
	utils.ProcessForm("/api/user/register", r.Data, r, model.UsernamePasswordValidator())
}

type PostView struct {
	PostId int
}

type ErrorListModel struct {
	Errors map[string]string
}

func main() {
	wade := wd.WadeUp("pg-home", "/web", "wade-content", "wpage-container", func(wade *wd.Wade) {
		// These things must go in here because they should be called at the right time when
		// the required HTML imports are all available.

		/* We call RegisterPages to register the pages and their associated route

		On the left are the url routes.
			A route may contain patterns like the ":postid" below.
			 | It describes a root parameter, for example /post/42 matches to
			 | the page "pg-post-view" with postid=42
		 On the right are page id's, they refer to the id of the <wpage> element,
		  | the id is unique to the page and it identifies the page.

		*/
		wade.Pager().RegisterPages(map[string]string{
			"/home":          "pg-home",
			"/posts":         "pg-post",
			"/posts/new":     "pg-post-new",
			"/post/:postid":  "pg-post-view",
			"/user":          "pg-user",
			"/user/login":    "pg-user-login",
			"/user/register": "pg-user-register",
			"/404":           "pg-not-found",
		})

		wade.Pager().SetNotFoundPage("pg-not-found")

		/* Register custom tags to be used in the html content

		The second parameters in these function calls, the "prototype"s
		are required so that Wade knows the datatype of the new
		custom element's attributes.
		The prototype must be a struct and not a pointer.
		It will be copied and new pointer instances will be made for each separate
		use of the custom element.

		*/
		wade.Custags().RegisterNew("t-userinfo", UserInfo{})
		wade.Custags().RegisterNew("t-errorlist", ErrorListModel{})
		wade.Custags().RegisterNew("t-test", UsernamePassword{})

		/* This sets the controller for the page "pg-user-login"
		The controller function returns a model, of which fields are used as targets
		for data binding in the page.
		In this case, "austat" is returned, and its AuthGened field is used
		for HTML bind-if to show whether the authentication info is generated
		or being generated
		*/
		wade.Pager().RegisterController("pg-user-login", func(p *wd.PageData) interface{} {
			req := http.Service().NewRequest(http.MethodGet, "/auth")
			austat := &AuthedStat{false}
			// performs the request to auth asynchronously
			ch := req.Do()

			// use a goroutine to process the response
			go func() {
				u := new(model.User)
				// here we wait for the response to come from the channel
				// and decode it to u
				(<-ch).DecodeDataTo(u)

				pdata.Service().Set("authToken", u.Token)

				// we set as.AuthGened to true here, the html elems that are bound
				// to this field will update accordingly
				austat.AuthGened = true
			}()
			return austat
		})

		// Too lazy to type this comment
		wade.Pager().RegisterController("pg-user-register", func(p *wd.PageData) interface{} {
			ureg := new(RegUser)
			// The RegUser struct contains a lot, please read the RegUser struct code
			// near to top to know more.

			/* This must be called for models that embed Validated for validation.

			It simply creates an entry in the Validated.Errors map
			for each field of ureg.Data.
			Without this we cannot use something like
			"Errors: Errors.Username" for the binding of a <t-errorlist>
			*/
			ureg.Validated.Init(ureg.Data)
			return ureg
		})

		// Too lazy to type this comment
		wade.Pager().RegisterController("pg-post-view", func(p *wd.PageData) interface{} {
			pv := new(PostView)
			// Remember the route parameter :postid above?
			// The call below puts its value into pv.PostId
			// so that if we visit page /post/42, pv.PostId becomes 42
			p.ExportParam("postid", &pv.PostId)
			return pv
		})
	})

	// This part adds a function to be called to modify every http request
	// It sets the AuthToken header to a token that will be verified by the server
	http.Service().AddHttpInterceptor(func(req *http.Request) {
		token, ok := pdata.Service().GetStr("authToken")
		if !ok {
			return
		}
		req.Headers.Set("AuthToken", token)
	})

	// Should must literally be called at the bottom of every Wade application
	// for whatever the reason
	wade.Start()
}

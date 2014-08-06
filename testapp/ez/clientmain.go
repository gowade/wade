package main

import (
	"strings"

	wd "github.com/phaikawl/wade"
	"github.com/phaikawl/wade/elements/menu"
	"github.com/phaikawl/wade/services/http"
	"github.com/phaikawl/wade/services/pdata"
	"github.com/phaikawl/wade/testapp/ez/model"
	"github.com/phaikawl/wade/utils"
)

type UserInfo struct {
	Name string
	Age  int
}

func (uinf *UserInfo) Init(ce *wd.CustomElem) error {
	ce.Contents.SetHtml(strings.Replace(
		strings.Replace(ce.Contents.Html(), "&lt;3", `<span style="color: crimson">♥</span>`, -1),
		":wink:", `<span style="color: DimGray">◕‿↼</span>`, -1))
	return nil
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
	go func() {
		utils.ProcessForm("/api/user/register", r.Data, r, model.UsernamePasswordValidator())
	}()
}

type PostView struct {
	PostId int
}

type ErrorListModel struct {
	Errors map[string]string
}

type HomeView struct{}

func (hv *HomeView) Highlight(word string) string {
	return ">> <strong>" + word + "<strong> <<"
}

func main() {
	wade := wd.WadeUp("pg-home", "/web", func(wade *wd.Wade) {
		wade.Pager().RegisterDisplayScopes(map[string]wd.DisplayScope{
			"pg-home":          wd.MakePage("/home", "Home"),
			"pg-user-bio":      wd.MakePage("/user/bio", "Bio"),
			"pg-user-secrets":  wd.MakePage("/user/secrets", "Secrets"),
			"pg-user-register": wd.MakePage("/user/register", "Register"),
			"pg-user-login":    wd.MakePage("/user/login", "Login"),
			"pg-post":          wd.MakePage("/post", "Posting"),
			"pg-post-view":     wd.MakePage("/post/view/:postid", "Viewing post %v"),
			"pg-not-found":     wd.MakePage("/404", "Page not found"),

			"grp-user-profile": wd.MakePageGroup("pg-user-bio", "pg-user-secrets"),
		})

		wade.Pager().SetNotFoundPage("pg-not-found")

		/* Register custom tags to be used in the html content.

		Each value in the map in these function calls are "prototype"s
		They are required so that Wade knows the datatype of the new
		custom element's attributes.
		It will be copied and new pointer instances will be made for each separate
		use of the custom element.
		*/

		wade.RegisterCustomTags("/public/elements.html", map[string]interface{}{
			"userinfo":  UserInfo{},
			"errorlist": ErrorListModel{},
			"test":      UsernamePassword{},
		})

		// Import the menu custom element from wade's packages
		wade.RegisterCustomTags("/public/menu.html", menu.Spec())

		/* This sets the controller for the page "pg-user-login"
		The controller function returns a model, of which fields are used as targets
		for data binding in the page.
		In this case, "austat" is returned, and its AuthGened field is used
		for HTML bind-if to show whether the authentication info is generated
		or being generated
		*/
		wade.Pager().RegisterController("pg-user-login", func(p *wd.PageCtrl) interface{} {
			req := http.Service().NewRequest(http.MethodGet, "/auth")
			austat := &AuthedStat{false}
			// performs the request to auth asynchronously

			// use a goroutine to process the response
			go func() {
				u := new(model.User)
				// We perform a request
				req.Do().DecodeDataTo(u)

				pdata.Service().Set("authToken", u.Token)

				// we set as.AuthGened to true here, the html elems that are bound
				// to this field will update accordingly
				austat.AuthGened = true
			}()
			return austat
		})

		// Too lazy to type this comment
		wade.Pager().RegisterController("pg-user-register", func(p *wd.PageCtrl) interface{} {
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
		wade.Pager().RegisterController("pg-post-view", func(p *wd.PageCtrl) interface{} {
			pv := new(PostView)
			// Remember the route parameter :postid above?
			// The call below puts its value into pv.PostId
			// so that if we visit page /post/42, pv.PostId becomes 42
			p.ExportParam("postid", &pv.PostId)
			return pv
		})

		wade.Pager().RegisterController("grp-user-profile", func(p *wd.PageCtrl) interface{} {
			return UserInfo{
				Name: "Rivr Perf. Nguyen",
				Age:  18,
			}
		})

		wade.Pager().RegisterController("pg-home", func(p *wd.PageCtrl) interface{} {
			return new(HomeView)
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

package main

import (
	"fmt"
	"strings"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/elements/menu"
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/testapp/ez/model"
	"github.com/phaikawl/wade/utils"
)

type UserInfo struct {
	Name string
	Age  int
}

func (uinf *UserInfo) Init(ce wade.CustomElem) error {
	ce.Contents.SetHtml(strings.Replace(
		strings.Replace(ce.Contents.Html(), "&lt;3", `<span style="color: crimson">♥</span>`, -1),
		":wink:", `<span style="color: DimGray">◕‿↼</span>`, -1))
	return nil
}

type AuthedStat struct {
	AuthGened bool
}

type UsernamePassword struct {
	wade.NoInit
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
		utils.ProcessForm(http.DefaultClient(), "/api/user/register", r.Data, r, model.UsernamePasswordValidator())
	}()
}

type PostView struct {
	wade.NoInit
	PostId int
}

type ErrorListModel struct {
	wade.NoInit
	Errors map[string]string
}

type HomeView struct {
	wade.NoInit
}

func (hv *HomeView) Highlight(word string) string {
	return ">> <strong>" + word + "<strong> <<"
}

func mainFn(r wade.Registration) {
	r.RegisterDisplayScopes([]wade.PageDesc{
		wade.MakePage("pg-home", "/home", "Home"),
		wade.MakePage("pg-user-bio", "/user/bio", "Bio"),
		wade.MakePage("pg-user-secrets", "/user/secrets", "Secrets"),
		wade.MakePage("pg-user-register", "/user/register", "Register"),
		wade.MakePage("pg-user-login", "/user/login", "Login"),
		wade.MakePage("pg-post", "/post", "Posting"),
		wade.MakePage("pg-post-view", "/post/view/:postid", "Viewing post %v"),
		wade.MakePage("pg-not-found", "/404", "Page not found"),
	}, []wade.PageGroupDesc{
		wade.MakePageGroup("grp-user-profile", "pg-user-bio", "pg-user-secrets"),
	})

	/* Register custom tags to be used in the html content.

	Each value in the map in these function calls are "prototype"s
	They are required so that Wade knows the datatype of the new
	custom element's attributes.
	It will be copied and new pointer instances will be made for each separate
	use of the custom element.
	*/

	r.RegisterCustomTags("/public/elements.html", map[string]wade.CustomElemProto{
		"userinfo":  &UserInfo{},
		"errorlist": &ErrorListModel{},
		"test":      &UsernamePassword{},
	})

	// Import the menu custom element from wade's packages
	r.RegisterCustomTags("/public/menu.html", menu.Spec())

	/* This sets the controller for the page "pg-user-login"
	The controller function returns a model, of which fields are used as targets
	for data binding in the page.
	In this case, "austat" is returned, and its AuthGened field is used
	for HTML bind-if to show whether the authentication info is generated
	or being generated
	*/
	r.RegisterController("pg-user-login", func(p wade.ThisPage) interface{} {
		austat := &AuthedStat{false}
		// performs the request to auth asynchronously

		go func() {
			resp, err := p.Services().Http.GET("/auth")
			if err != nil || resp.Failed() {
				return
			}

			// we set as.AuthGened to true here, the html elems that are bound
			// to this field will update accordingly
			austat.AuthGened = true
		}()
		return austat
	})

	// Too lazy to type this comment
	r.RegisterController("pg-user-register", func(p wade.ThisPage) interface{} {
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
	r.RegisterController("pg-post-view", func(p wade.ThisPage) interface{} {
		pv := new(PostView)
		// Remember the route parameter :postid above?
		// The call below puts its value into pv.PostId
		// so that if we visit page /post/42, pv.PostId becomes 42
		p.GetParam("postid", &pv.PostId)
		return pv
	})

	r.RegisterController("grp-user-profile", func(p wade.ThisPage) interface{} {
		return UserInfo{
			Name: "Rivr Perf. Nguyen",
			Age:  18,
		}
	})

	r.RegisterController("pg-home", func(p wade.ThisPage) interface{} {
		return new(HomeView)
	})
}

func main() {
	err := wade.StartApp(wade.AppConfig{
		StartPage: "pg-home",
		BasePath:  "/web",
	}, mainFn)
	if err != nil {
		panic(fmt.Errorf("Failed to load with error %v", err))
	}
}

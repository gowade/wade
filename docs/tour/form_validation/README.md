# Form validation

In the previous section, you probably have wondered where does `Errors.Username` come from and where is it in the page model. Theres no magic.

Our model for the page `pg-user-register` is in `clientmain.go`:

    type RegUser struct {
    	utils.Validated
    	Data UsernamePassword
    }

RegUser embeds `utils.Validated`

    type ErrorMap map[string]map[string]string
    type Validated struct {
    	Errors ErrorMap
    }
That's why we can access an `Errors` field in the page model. But why does `Errors` have a `Username` field? Why does `Errors.Username` work?

Inside `pg-user-register`'s controller func, `ureq.Data` is

    type UsernamePassword struct {
    	Username string
    	Password string
    }
so the call to [Validated.Init](http://godoc.org/github.com/phaikawl/wade/utils#Validated.Init)

    ureg.Validated.Init(ureg.Data)
sets `Validated.Errors` to

    map[string]map[string]string {
        "Username": map[string]string {
        },
        "Password": map[string]string {
        },
    }
That's why there's a "Username" field.

More over, `Validated.Errors` is also updated whenever we click the "Submit" button

    <a bind-on-click="Submit">Submit</a>
which calls `RegUser.Submit`, which in turn calls Wade's [utils.ProcessForm](http://godoc.org/github.com/phaikawl/wade/utils#ProcessForm).

    func (r *RegUser) Submit() {
    	utils.ProcessForm("/api/user/register", r.Data, r, model.UsernamePasswordValidator())
    }
(Please read the reference link above to know how ProcessForm works)

Hence `Errors.Username` a.k.a `Validated.Errors.Username`, could be accessed from the HTML template, it is updated whenever `Submit` is clicked, and the output error list is automatically updated thanks to data binding.

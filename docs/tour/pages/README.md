# Pages

You should take a look at `public/pages.html` to have an idea how pages are laid out. It contains a hierarchy of `<wpage>` elements. Each `<wpage>` describes contents of a web page and its children. Elements are shared between pages of a common ancestor.

Yes we can describe a whole bunch of different pages, like Post View, User Profile, etc. inside a single HTML! (If something is too big, too long we can split it into a separate file and import with a `wimport`, _ez_).

Let's take `pg-user` as an example.

    <wpage id="pg-user" title="User page">
    	<wpage id="pg-user-login">
    		<span bind-if="AuthGened">
    		    Authentication info is generated.
    		</span>
    		<span bind-ifn="AuthGened">
    		    Generating auth info...
    		</span>
    	</wpage>
    	<wpage id="pg-user-profile">
            Here is the profile of the most awesome person in the world:
        </wpage>
        <t-userinfo name="Awesome" age="99"></t-userinfo>
    </wpage>
Here we have the parent page `pg-user` with its 2 children `pg-user-login` and `pg-user-profile`.

The `t-userinfo` is shared between those two children pages so for example if user accesses `/web/user/login` (the url of `pg-user-login` as registered in `clientmain.go`), something like this is displayed:

    Awesome, 99
    Generating auth info...

Each `<wpage>` must have a unique id, and must be registered with [RegisterPages](http://godoc.org/github.com/phaikawl/wade#PageManager.RegisterPages). We have this inside `clientmain.go`:

    wade.Pager().RegisterPages(map[string]string{
			"/home":          "pg-home",
			"/posts":         "pg-post",
			"/posts/new":     "pg-post-new",
			"/post/:postid":  "pg-post-view",
			"/user":          "pg-user",
			"/user/login":    "pg-user-login",
			"/user/profile":  "pg-user-profile",
			"/user/register": "pg-user-register",
			"/404":           "pg-not-found",
		})
It's recommended that page ids have prefix `pg-` to avoid confusion.

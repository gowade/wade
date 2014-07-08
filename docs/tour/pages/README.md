# Pages

Let's take a look at `public/pages.html` to have an idea how pages are laid out. It contains a hierarchy of `<wpage>` elements. Each `<wpage>` describes contents of a web page and its children. Elements are shared between pages of a common ancestor.

Yes we can describe a whole bunch of different pages, like Post View, User Profile, etc. inside a single HTML! (If something is too big, too long we can easily put it into a separate HTML file and import with a `wimport`).

Let's take `pg-user` as an example.

    <wpage pid="pg-user" route="/user" title="User page">

		<wpage pid="pg-user-login" route="/user/login">
			<span bind-if="AuthGened">Authentication info is generated.</span>
			<span bind-ifn="AuthGened">Generating auth info...</span>
		</wpage>

		<wpage pid="pg-user-profile" route="/user/profile">
			Here is the profile of the most awesome person in the world:
		</wpage>

		<userinfo name="Awesome" age="99"></userinfo>
    </wpage>
Here we have the parent page `pg-user` with its 2 children `pg-user-login` and `pg-user-profile`.

Each page has a `pid` attribute which is its **unique** page ID, a `route` attribute for its url route, and a `title`. Note that the path for `route` is absolute, it does not care about the hierarchy.

The `<userinfo>` element at the end is shared between `pg-user-login` and `pg-user-profile`. So for example if user accesses `/web/user/login`, something like this is displayed:

    Generating auth info...
    Awesome, 99

##Registering

We have to call [RegisterPages](http://godoc.org/github.com/phaikawl/wade#PageManager.RegisterPages) in `clientmain.go` to register all the pages in `pages.html`, otherwise they are not recognized.

Inside `index.html`,

	<script type="text/wadin">
		<wpage id="wpage-root">
    		<wimport src="/public/pages.html"></wimport>
		</wpage>
		<wimport src="/public/elements.html"></wimport>
	</script>

we have `#wpage-root` as the root ancestor of all pages inside `pages.html`. So we register the pages like this

    wade.Pager().RegisterPages("wpage-root")

Keep in mind that the element used as root must be a `wpage`.

That's it!

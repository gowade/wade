# Pages

Let's take a look at `public/pages.html` to have an idea how pages are laid out. It contains a hierarchy of `<wpage>` elements. Each `<wpage>` describes contents of a web page and its children. Elements are shared between pages of a common ancestor.

Yes we can describe a whole bunch of different pages, like Post View, User Profile, etc. inside a single HTML! (If something is too big, too long we can easily put it into a separate HTML file and import with a `wimport`).

Let's take `pg-user-profile` as an example. It's declared in `pages.html`, and it's contents is in `user_profile.html`. Here's a cut out version to illustrate.

    <wpage pid="pg-user-profile" route="/user/profile" title="User profile">
        <userinfo bind="Name: Name; Age: Age"></userinfo>

        <wpage pid="pg-user-bio" route="/user/bio">
           	<wrep target="menu">[...]</wrep>

            This girl is awesome...
        </wpage>

        <wpage pid="pg-user-secrets" route="/user/secrets">
            <wrep target="menu">[...] </wrep>

            <div>Confidental information, revealed:</div>
            [...]
        </wpage>
    </wpage>

Here we have the parent page `pg-user-profile` with its 2 children `pg-user-bio` and `pg-user-secrets`.

Each page has a `pid` attribute which is its **unique** page ID, a `route` attribute for its url route, and a `title`. Note that the path for `route` is absolute, it does not care about the hierarchy.

The `<userinfo>` at the beginning is shared between `pg-user-bio` and `pg-user-secrets`. So when the user accesses `/web/user/bio`, something like this is displayed:

    <userinfo>Rivr Perf. Nguyen, 18</userinfo>
    This girl is awesome...

For `/web/user/secrets`:
    <userinfo>Rivr Perf. Nguyen, 18</userinfo>
    <div>Confidental information, revealed:</div>
    [...]

The parent `pg-user-profile` is also a page with route, so for it, something like this is displayed:

    <userinfo>Rivr Perf. Nguyen, 18</userinfo>
Yeah that is our common `<userinfo>` element declared in `pg-user-profile`.

##Sections
There is often a need to use a common layout, for example we want a layout with a menu and a main content container. We could put the `<wpage>` hierarchy in the main content container, but where do we put the menus for each different page? That's when we make use of `<wsection>`.

Let's look at how it it used in `pages.html`.

We use the `div.header` as the container for the (common) upper part of pages. And the `div.jumbotron` (our main content container) is our main parent of all those `<wpage>`'s `pg-home`, `pg-user-profile` etc.

    <div class="header">
	    <ul class="nav nav-pills pull-right" role="tablist">
            <wsection name="menu">
                <li><a bind-page="url(`pg-home`)">Home</a></li>
            </wsection>
	    </ul>
	    <h3 class="text-muted">Wade.go Demo</h3>
    </div>

    <div class="jumbotron">
        <wpage pid="pg-home"...>
        </wpage>
        <wpage pid="pg-user-profile">
        </wpage..
        ...]
    </div>

The `<wsection>` element is somewhat like a placeholder with a name, in this case it's `menu`. It has some default contents (above it's the link to home page), and it could be replaced by a later use of `<wrep` (wreplace).

For example the home page we have

    <wpage pid="pg-home" route="/home" title="Wade Home">
        <wrep target="menu">
            <li class="active"><a bind-page="url(`pg-home`)">Home</a></li>
            <li><a bind-page="url(`pg-post`)">Posting</a></li>
            <li><a bind-page="url(`pg-user-profile`)">Profile</a></li>
            <li><a bind-page="url(`pg-user-login`)">Login</a></li>
            <li><a bind-page="url(`pg-user-register`)">Register</a></li>
        </wrep>
        Welcome to this world, a world of awesomeness.
    </wpage>

So the default contents of the `menu` section (the link to home) is replaced with the contents of this "menu" `wrep`. That's why in `pg-home` we have a 5 links and the `home` link displayed as `.active`.

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

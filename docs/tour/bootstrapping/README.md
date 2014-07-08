# The bootstrap

The app starts in the *main()* function of `testapp/ez/clientmain.go`. There are a lot of comments describing the calls inside, please read those along the way.


small reminder: there will be a lot of blue links for function names or other things, those are links to the the [API reference](http://godoc.org/github.com/phaikawl/wade) (not ads obviously), so please clink on them for details.

The [WadeUp](http://godoc.org/github.com/phaikawl/wade#WadeUp) call

    wade := wd.WadeUp("pg-home", "/web", func(wade *wd.Wade)...
initializes the app with `/web` as the app's base path and "pg-home" as the starting page. It acceps the third parameter as a function that will be called at the right time after initialization and HTML imports.

##HTML

`index.html` is the master HTML, when accessing the app's urls, the server code always just returns this file. Everything is rendered by Wade's client code.

In the file you can see a funny `script` element with type "text/wadin":

    <script type="text/wadin">
		<wpage id="wpage-root">
    		<wimport src="/public/pages.html"></wimport>
		</wpage>
		<wimport src="/public/elements.html"></wimport>
	</script>
It is the container for all your HTML source code, it's not displayed and is completely ignored by the browser, screen readers, etc. Its HTML content will be processed and copied to a real element by Wade. The code above imports (actually just like "include") 2 files:
* `public/pages.html`, for marking up pages
* `public/elements.html`, for declaring some [custom elements]().


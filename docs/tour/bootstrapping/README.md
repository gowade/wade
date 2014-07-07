# The bootstrap

The app starts in the *main()* function of `testapp/ez/clientmain.go`. There are a lot of comments describing the calls inside, please read those along the way.


small reminder: there will be a lot of blue links for function names or other things, those are links to the the [API reference](http://godoc.org/github.com/phaikawl/wade) (not ads obviously), so please clink on them for details.

The [WadeUp](http://godoc.org/github.com/phaikawl/wade#WadeUp) call

    wd.WadeUp("pg-home", "/web", "wade-content", "wpage-container", func)
initializes the app with `/web` as the app's base path and some parameters related to things in the HTML template.

`index.html` is the master template. In the file you can see the funny `#wade-content`

    <script type="text/wadin" id="wade-content">
        <wimport src="/public/pages.html"></wimport>
        <wimport src="/public/elements.html"></wimport>
    </script>
which is the container for all your HTML source (or template) code, it's a `script` element, so it's not displayed and is ignored by the browser and screen readers. The code above imports (actually just a simple "include") 2 files:
* `public/pages.html`, for marking up pages
* `public/elements.html`, for declaring some [custom elements]().

There's also the `#wpage-container`

    <div id="wpage-container">
which is where all the real *rendered* HTML for the browser to display is put into. So of course it should be empty.


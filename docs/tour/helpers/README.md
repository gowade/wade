# Helpers

You saw these inside `pages.html` right?

    <a bind-page="url(`pg-home`)">Home</a>
    <a bind-page="url(`pg-post-view`, 42)">
        Post #42: Life, universe and everything.
    </a>

What magic is this? You probably got it already, these look like function calls, and in fact they are function calls.

`url` is a core helper that returns the url info for a specific page, used with [bind-page](http://godoc.org/github.com/phaikawl/wade/bind#PageBinder) for making page links. Its first parameter is a page id, the following ones are *route parameters*.

The page `pg-post-view` has route `"/post/view/:postid"`. So the second call above returns `"/post/view/42"` (`postid` is a *route parameter*).

As you can see, we can use string literals (must be wrapped with backticks `` ` instead of quotes) and numbers in the syntax.

You can even call other helpers and do something silly like this for the sake of... trolling

    <a bind-page="url(concat(`pg-`, concat(`post-`, `view`)), 42)">
        Post #42: Life, universe and everything.
    </a>

But that's all, our parser is very strict but very dumb, it doesn't (and actually never will) understand any more advanced syntax, like operators. The code above is "too smart to handle" already, please don't write something like that in real code.

Custom helpers need to be registered before use. Global helpers are registered with [Binding.RegisterHelper](http://godoc.org/github.com/phaikawl/wade/bind#Binding.RegisterHelper), and local helpers (exist inside a page controller, only used for that page) are registered with [PageData.RegisterHelper](http://godoc.org/github.com/phaikawl/wade#PageData.RegisterHelper).

You can read the code of `bind/helpers.go` to see what default helpers are available.

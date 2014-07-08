# Basic data binding

Data binding is the standard mechanism in popular Javascript frameworks for connecting HTML and Javascript. It's awesome to have the HTML markup automatically updated whenever some data in our code changes. This functionality is a big focus for Wade.

You can take a look at the various usages of data binding in `pages.html`. All those `bind-something`, you already know what they mean, don't you?

Let's examine the `pg-user-login` page.

	<wpage pid="pg-user-login" route="/user/login">
		<span bind-if="AuthGened">Authentication info is generated.</span>
		<span bind-ifn="AuthGened">Generating auth info...</span>
	</wpage>

What is it supposed to do? It displays "Generating..." when we start sending a request to the server (to get authentication info) and changes to "generated" when that operation completes.

Inside `clientmain.go` we have a model struct

    type AuthedStat struct {
    	AuthGened bool
    }
and the call (cut out version)

    wade.Pager().RegisterController("pg-user-login",
    [...]
    	req := http.Service().NewRequest(http.MethodGet, "/auth")
    	austat := &AuthedStat{false}
    	responseChannel := req.Do()

    	// use a goroutine to process the response
    	go func() {
            // wait for the response
    		<-responseChannel
            // then set AuthGened to true
            austat.AuthGened = true
        }()
        return austat
    })

Here `austat`, returned by the controller function, is our *model* for the page. It is an `AuthedStat` that contains a bool field `AuthGened`, we use it to indicate whether the request is completed or not yet.

In the HTML code of `pg-user-login`
* [bind-if](http://godoc.org/github.com/phaikawl/wade/bind#IfBinder) is a [*dom binder*]() that displays the element when the referred value is true and hides it when the value is false.
* [bind-ifn](http://godoc.org/github.com/phaikawl/wade/bind#UnlessBinder) is the reverse (if-not)

So it should be clear to you how it works now. Basically first `austat.AuthGened` is set to `false`. The output shows for the `bind-ifn`

    Generating auth info...
We make a http request, and as soon as we get the response from the channel, we set `austat.AuthGened` to `true`. The HTML automatically updates and now shows for the `bind-if` instead

    Authentication info is generated.

Yes, that's it!

Inside `pages.html` there are also usages of other core *dom binders* including
* bind-html ([HtmlBinder](http://godoc.org/github.com/phaikawl/wade/bind#HtmlBinder))
* bind-value ([ValueBinder](http://godoc.org/github.com/phaikawl/wade/bind#ValueBinder))
* bind-attr ([AttrBinder](http://godoc.org/github.com/phaikawl/wade/bind#AttrBinder))
* bind-each ([EachBinder](http://godoc.org/github.com/phaikawl/wade/bind#EachBinder))
* bind-on ([EventBinder](http://godoc.org/github.com/phaikawl/wade/bind#EventBinder))

Detailed references are in the links and in `pages.html` itself, feel free to examine and play with them.


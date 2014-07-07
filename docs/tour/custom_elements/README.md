# Custom HTML tags

Wade.go supports a simple and effective mechanism for HTML code reuse, something like the future [HTML custom elements](http://www.html5rocks.com/en/tutorials/webcomponents/customelements/), but much simpler.

You must have noticed these things already in `pages.html`

    <userinfo name="Hai Thanh Nguyen" age="18"></userinfo>
    <userinfo name="Awesome" age="99"></userinfo>
    <errorlist bind="Errors: Errors.Username"></errorlist>

The meanings of these custom html tags are declared in `elements.html`.

For `userinfo`, we have

    <welement id="t-userinfo" attributes="Name Age">
        <p>
            <strong><% Name %></strong>,
            <em><% Age %></em>
        </p>
    </welement>
It has 2 public html attributes `name` and `age` specified with `attributes="Name Age"` (space-separated). These attributes must be capitalized to match exactly with the exported field names of a prototype struct, which is `UserInfo` inside `clientmain.go`:

    type UserInfo struct {
    	Name string
    	Age  int
    }
The prototype struct describes the datatypes of those html attributes. And the custom tag is registered inside `clientmain.go` with a call to [Custags().RegisterNew](http://godoc.org/github.com/phaikawl/wade#Wade.Custags)

    wade.Custags().RegisterNew("userinfo", UserInfo{})

That's all required to "invent" our own `userinfo` html tag.

So we can now use it like this in `pages.html`:

    <userinfo name="Hai Thanh Nguyen" age="18"></userinfo>
Wade assigns the attributes and puts the real contents from `userinfo` into the element, so after processing, the displayed thing looks like this:

    <userinfo name="Hai Thanh Nguyen" age="18">
        <p>
            <strong>Hai Thanh Nguyen</strong>,
            <em>18</em>
        </p>
    </userinfo>
Custom tags must be registered to be used, Wade.go doesn't "magically" detect them. Also keep in mind that a separate copy of the prototype struct is created for each separate use of the custom tag.

##Attribute binding
What if we want to pass complex data from the page model into the contents of a custom tag? How to define a tag to display a list of validation errors could be used for boh Username's error list and Password's error list or anything?

Here's the register call in `clientmain.go`:

    wade.Custags().RegisterNew("errorlist", "t-errorlist", ErrorListModel{})

So `<errorlist>` has ErrorListModel as the prototype, which is

    type ErrorListModel struct {
    	Errors map[string]string
    }

Let's look at how `errorlist` is used in `pg-user-register`.

    <errorlist bind="Errors: Errors.Username"></errorlist>
The syntax is straightforward, it means we bind the value of `Errors.Username` in the page's model, to the attribute `Errors` of an errorlist's `ErrorListModel` instance.

`Errors.Username` is a `map[string]string` which holds the list of validation errors for the Username. (The question about where it is inside the page model struct is answered in the next section, not important for now)

    <welement id="t-errorlist" attributes="Errors">
        <div class="error">
    		<ul>
    			<li bind-each="Errors -> key, error">
    				error type #<% key %>
    				<% error %>
    			</li>
    		</ul>
    	</div>
    </welement>

So our custom element's `Errors` attribute is bound to the page's `Errors.Username`, which is a map. We use `bind-each` to make a loop through `Errors`. After `->` are *outputs*, it is Wade's explicit way to name the things that would be bound to elements inside.

Generally, the [each binder](http://godoc.org/github.com/phaikawl/wade/bind#EachBinder) emits 2 outputs, a key and a value, just like when you're looping through a map with `range` in Go. Here for each item, the key is bound to `key`, and the value is bound to `error`. You can see how we access those values with `<% key %>` and `<% error %>` inside.

This will effectively list all those errors and the result could be something like:

    <ul>
		<li>
			error type #minChar
			Not enough characters.
		</li>
		<li>
			error type #zero
			Must not be empty
		</li>
    </ul>




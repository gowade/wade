Wade.go [![godoc reference](http://b.repl.ca/v1/godoc-reference-brightgreen.png)](http://godoc.org/github.com/phaikawl/wade)
====
The no-magic client-side web framework for Go->js.  

  
Instructions on running the companion app Brogpal: https://github.com/phaikawl/brogpal.  

#How it works
##The flow
The server simply returns the `index.html` everytime the user visits the site, without any template/rendering on server side, the client controls the whole flow, render and direct the pages, the server is just a resource manager, an API provider, which returns needed resources (for example user info from the database) on ajax requests of the client, which is written in Go and compiled into Javascript. 

In the companion app, `index.html` is the root html, it imports `pages.html`, which defines the pages, and `elements.html`, which defines custom elements for the site.


##Html import
Html import is a simple mechanism that allows splitting code into multiple files. It simply replaces the `<wimport>` element with the html file's content, no magic whatsoever.
Usage:

    <wimport src="some_path"></wimport>

##Pages
Each page is declared with a `wpage` element, the elements in the page are put inside of those tags. Pages need to be registered with `wade.Pager().RegisterPages()` in the client code. Wade uses HTML5 history to save page states and handle brower Forward/Back buttons.
Within the hierachy, pages can share elements with each other very easily without the need for any kind of inheritance or import.

##Data binding
###Dom binding
Wade has support for data binding between HTML and Go/Js models, it's called *Dom binding*. 
 
For each page, a *page controller* could be registered with `wade.Pager().RegisterController`, which will be called every time the specified page loads to control the page. Each page controller returns a model, which will be bound to the whole page.
If we have a struct

    type UserReg struct {
      Data struct {
        Username string
        Password string
      }
    }

The page handler:
    
    wade.Pager().RegisterHandler("pg-user-register", func() interface{} {
      return new(UserReg)
    })

The returned UserReg instance will be bound to the page, and for example within the page *pg-user-register*, we have something like this:

    Username: <input type="text" bind-value="Data.Username"/>

The *value* attribute (the text) of the input field will be bound/synchronized with the model's Data.Username field so that when any change happens on one side, the other side will update according to that.

In the example above, we used the *Value binder*, which is one of the default binders declared in *binders.go*. Some other basic dom binders:
* **bind-html**: binds to the html content of the element
* **bind-attr**-*someattr*: binds to the attribute *someattr* of the element

####The each binder
**bind-each** is a dom binder with 2 outputs *key*, *value* for each item. Example usage:

    <ul>
		<li bind-each="Errors -> type, msg"><p>
		    Type: <span bind-html="type"></span>,
			Message: "<span bind-html="msg"></span>"
		</p></li>
	</ul>

Here for each element of Errors, the key is bound to *type* and the value is bound to *msg* for the things inside the element.  
If Errors is currently

    map[string]string {
        "minChar": "Not enough characters.",
        "invalidChar": "Invalid characters.",
    }

The output will be like

    Type: minChar, Message: "Not enough characters."
    Type: invalidChar, Message: "Invalid characters."

####General usage for dom binding:
**bind-**<*binder_name*>[-*arg1, arg2,*...] = "*target* [-> *output1, output2,...*]" 

"-> *output1*, *output2*..." is the syntax for explicitly specifying the binder's *outputs*, which will be bound to the things inside the element. The *each* binder above is an example of this, it emits 2 output.

Each dash argument *arg1*, *arg2*... above is passed into the binder's functions as a list of strings. For example the **bind-attr** binder is used like this:

    <a bind-attr-href="Url">Wade</a>
It binds the attribute specified after **bind-attr-**, which in this case is *href* to *Url*.

##Custom elements
Custom elements can be declared with `welement` and registered with `wade.RegisterNewTag()`. This is useful for HTML code reuse.  
For example, we define a custom element tag called *t-userinfo*:  

    <welement id="t-userinfo" attributes="Name Country">
        <p>
            <strong bind-html="Name"></strong>,
            <em bind-html="Country"></em>
        </p>
    </welement>

It's considered a html tag with attributes `attr-Name` and `attr-Country` now, we can use it like this:

    <t-userinfo attr-Name="Hai Thanh Nguyen" attr-Country="Vietnam"></t-userinfo>
Each custom element is bound to a unique model, which holds the datatype and value for attributes of that element.

###Custom element attribute binding
There are times when we need to bind a custom element's attribute to a data member of the page's model.  
Here's a real world example, we want to declare a custom element for displaying form errors.

    <welement id="t-errorlist" attributes="Errors">
        <div class="error">
    		<ul>
    			<li bind-each="Errors -> _, msg">
    				<span bind-html="msg"></span>
    			</li>
    		</ul>
    	</div>
    </welement>


#Shoutouts to
* Richard Musiol [Gopherjs](http://github.com/gopherjs/gopherjs) for what makes Wade possible
* [Rivets.js](http://rivetsjs.com) for the awesome binding mechanism, which Wade is heavily heavily inspired by (or based on)
* [Polymer](http://polymer-project.org) and [Pager.js](http://pagerjs.com) for ideas on page element and custom elements
* [Watch.js](https://github.com/melanke/Watch.JS)

#License

    Copyright (c) 2014, Hai Thanh Nguyen
    All rights reserved.

    Redistribution and use in source and binary forms, with or without
    modification, are permitted provided that the following conditions are met:
        * Redistributions of source code must retain the above copyright
          notice, this list of conditions and the following disclaimer.
        * Redistributions in binary form must reproduce the above copyright
          notice, this list of conditions and the following disclaimer in the
          documentation and/or other materials provided with the distribution.
        * Neither the name of VosDev nor the
          names of this software's contributors may be used to endorse or promote products
          derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
    ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
    WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
    DISCLAIMED. IN NO EVENT SHALL HAI THANH NGUYEN BE LIABLE FOR ANY
    DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
    (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
    LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
    ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
    (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
    SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

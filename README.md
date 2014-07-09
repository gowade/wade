Wade.go [![godoc reference](http://b.repl.ca/v1/godoc-reference-brightgreen.png)](http://godoc.org/github.com/phaikawl/wade) [![tutorial ready](http://b.repl.ca/v1/tutorial-ready-brightgreen.png)](http://phaikawl.gitbooks.io/wa-de-go-the-tour/)
====
The comprehensive client-side web framework for Go -> js.

*Brought to you with love and excitement...*
#A brand new way of developing
No it wasn't a typo, Wade is a fresh **client-side** web framework for [**Go**](http://golang.org), created to bring these together:
* The awesome convenience of client-side Web development with HTML data binding (think AngularJS)
* The Go platform. The true Go-style concurrency syntax. And the advantages of coding in a compiled programming language without having to sacrifice productivity.
* The awesome feeling of being able to use the same language for both client and server.

#####How is that even possible?  
Wade.go is compiled with [gopherjs](https://github.com/gopherjs/gopherjs), a Go -> Javascript transpiler. Gopherjs has evolved to become an awesome platform for writing client-side web application in Go. It provides strong, convenient Go-Javascript interoperability, compatibility with most of the Go standard library, and full goroutine support. However, there's no good existing web framework/wrapper for it yet, that's why Wade is created.

#Features
* Developed from wonderful ideas in existing Javascript libraries.
    * Data binding inspired by [Rivets.js](http://rivetsjs.com)'s awesome mechanism
    * Custom elements inspired by [Polymer](http://polymer-project.org), but much simpler
    * Convenient pages declaration inspired by [Pager.js](http://pagerjs.com)  
* Type safe, strict and defensive, designed for Go from the ground up, Wade takes full advantage of the type system

#Markup overview
Below are examples to show how working with HTML in Wade.go looks like.

An example Register page:

    <wpage pid="pg-user-register" route="/user/register" title="Register">
    
		Username:
		<input type="text" bind-value="Data.Username"></input>
		<errorlist bind="Errors: Errors.Username"></errorlist>
		
		Password:
		<input type="password" bind-value="Data.Password"></input>
		<errorlist bind="Errors: Errors.Password"></errorlist>
		
		<button bind-on-click="Reset">Reset</button>
		<button bind-on-click="Submit">Submit</button>
		
	</wpage>

That's how we create `<input>` fields with all the data automatically updated for `Username` and `Password`.

A method named `Reset` is called when the Reset button is clicked, similarly for Submit.

The `errorlist` above is a custom element, which is declared as

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

In the form, `Errors.Username` is a list of validation errors for the Username field. You can see that we can comfortably bind `errorlist`'s Errors attribute to a whole list that is `Errors.Username`, or `Errors.Password`. The elements inside have access to the list via `Errors`, we use a simple `bind-each` to loop through each error and display a list to the user.

Does it look good to you? So...

#Let's get started!
* Quick introduction: [Wa.de.Go! The Tour](http://phaikawl.gitbooks.io/wa-de-go-the-tour/).
* [Concepts reference]() (Coming soon)
* [API reference](http://godoc.org/github.com/phaikawl/wade)
    * [Binders reference ](http://godoc.org/github.com/phaikawl/wade/bind)

#Is it ready?
**Yes** it's ready to be tested, but **No** it's not yet ready for a public announcement, not well tested, still needs some feedbacks and refinements, a lot of stability problems may arise.  
Basically, "closed beta" for now.

#Contributing
Pull requests are welcome. Wade is young, feedbacks are necessary, the core functionalities are there but lot of things could be developed, like a (separate) package for authorization, websocket integration, etc...

[TODO list](https://github.com/phaikawl/wade/wiki/TODO).

Feel free to contact me at https://plus.google.com/+HaiThanhNguyenPk if you're interested.

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

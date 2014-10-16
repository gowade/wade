
>>> I'm happy to announce that Wade.Go is now in **Beta Testing** phase, from now on it's unlikely that the API and binding syntax will be changed drastically. No release yet though, the *iteration 3* is still being developed that brings necessary performance improvements

# Wade.Go
**Wade.Go** is a **client-centric** web framework like nothing you ever heard. It brings these awesome things together
* Compiled, statically typed programming with Go (compiled to Javascript on client side by [gopherjs](https://github.com/gopherjs/gopherjs))
* Client-centric web development model with HTML data binding (think AngularJs)
* Hybrid rendering: Write code once, render on both client and server (think server-side ReactJS)
* Instant functional testing with native `go test` (no browser needed)

With the creation of Wade.Go, you can now write both client and server in a single programming language that is not Javascript. Go brings the best concurrency pattern (goroutines) and the static type system that makes maintenance a breeze.

Although being a client-centric web framework, Wade.Go works even when Javascript is disabled and is SEO-friendly due to the ability to render the site with server-side Go. Wade.Go is built for web *sites*, not just *apps*, it works for content-heavy sites like blog, forum as well as very dynamic ones like Facebook.

# Markup intro
## Data binding
Wade.Go's templating is HTML-based, the data binding mechanism is partly inspired by a design document from Angular 2.0 and Rivetsjs. Features
  * Concise, strict and clear syntax, composed of a few clearly defined rules, no surprises!
  * Optional watching: you can control what data is watched for changes and what don't need

A little example

    <!-- '$' means "watch this value" -->
    <div #each(_,post)="$Posts">
        <div>
          <a @href="GetLink(post)">{{ post.Title }}</a>
        </div>
        <div>
            {{ len($post.Comments) }} comments
        </div>
        <div>
            by {{ post.Author }}
        </div>
    </div>
    
## Reusable Components
This feature is inspired by ReactJs and HTML Web Components. It's even better with the static type system from Go, prototype fields are all properly typed!

Example component prototype

    // VoteBoxModel is the prototype for the "votebox" custom element
    VoteBoxModel struct {
        *wade.BaseProto
        Vote      *c.Score
        VoteUrl   string
    }
    
Example usage and property binding

    <div #each(_,post)="$Posts">
      <div>
        <!-- Here we assign value to the fields Vote and VoteUrl of the component instance-->
        <VoteBox @Vote="post.Vote" @VoteUrl="GetVoteUrl(post)"></VoteBox>
      </div>
      <div>
        {{ post.Content }}
      </div>
    </div>

# Getting started
* [Tutorial](https://github.com/phaikawl/wade/wiki/Wade.Go-Quick-Start-Guide)
* [godoc API reference](http://godoc.org/github.com/phaikawl/wade)
* [Wadereddi](https://github.com/phaikawl/wadereddi) the demo app

Wade.Go is currently in "Beta testing" phase, nothing is really finalized yet, so please give feedbacks via the [Issues](https://github.com/phaikawl/wade/issues) section, tell us what you think about the framework, what you want to change and improve. You can also post [here](https://groups.google.com/forum/#!forum/wadego).

Pull requests are welcome. Contact me if you want to join the team.

# License

Copyright (c) 2014, Hai Thanh Nguyen
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.






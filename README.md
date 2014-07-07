Wade.go [![godoc reference](http://b.repl.ca/v1/godoc-reference-brightgreen.png)](http://godoc.org/github.com/phaikawl/wade)
====
The comprehensive client-side web framework for Go -> js.

*Brought to you with love and excitement...*
#A brand new way of developing
No it wasn't a typo, Wade is a fresh **client-side** web framework for **Go**, created to bring these together:
* The awesome convenience of client-side Web development with HTML data binding (think AngularJS)
* The Go platform. The true Go-style concurrency syntax. And the advantages of coding in a compiled programming language without having to sacrifice productivity.
* The awesome feeling of being able to use the same language for both client and server.

#####How is that even possible?  
Wade.go is compiled with [gopherjs](https://github.com/gopherjs/gopherjs), a Go -> Javascript transpiler. Gopherjs has evolved to become an awesome platform for writing client-side web application in Go. It provides strong, convenient Go-Javascript interoperability, compatibility with most of the Go standard library, and full goroutine support. However, there's no good existing web framework/wrapper for it yet, that's why Wade is created.

#Features
    
* *Explicit*, *Defensive* and *Strict*:  
The user code knows what's going on, and whenever there are mistakes like a wrong data binding syntax or refering a non-existent model field in the HTML, Wade raises a descriptive error message instead of silently ignoring.

* *Type safe*:  
Although being compiled to Javascript under the hood, Wade is designed for Go from the ground up and take full advantage of the type system.

Developed from wonderful ideas of existing Javascript libraries.
* Data binding mechanism inspired by [Rivets.js](http://rivetsjs.com)'s beautiful and customizable one
* Custom elements, inspired by [Polymer](http://polymer-project.org), but much simpler
* Convenient web page declaration inspired by [Pager.js](http://pagerjs.com)  

#Let's get started!
* Start here: [Wa.de.Go! The Tour](http://phaikawl.gitbooks.io/wa-de-go-the-tour/).
* [Concepts reference]() (Coming soon)
* [API reference](http://godoc.org/github.com/phaikawl/wade)
    * [Binders reference ](http://godoc.org/github.com/phaikawl/wade/bind)

#Is it ready?
**Yes** it's ready to be used! But **No** it's not yet ready for a public announcement, not well tested, still needs some feedbacks and refinements, a lot of stability problems may arise.  
Basically, "closed beta" for now.

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

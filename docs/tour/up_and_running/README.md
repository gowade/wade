# Up and running

We gotta run the *ez* demo app, it's ez. Let's get started!

##Prerequisites
First you need the following things
* [Go 1.3](http://golang.org/doc/install) with system PATH properly set up
(older versions of Go will **NOT** work!)
* Latest [fresh](https://github.com/pilu/fresh) and [gopherjs](https://github.com/gopherjs/gopherjs) installed and working as commands.
Here's how to install them:
    * `go get -u github.com/gopherjs/gopherjs`
    * `go get -u github.com/pilu/fresh`
* [bower](http://bower.io) installed and working as a command

##Installing
* Install [Wade.go](https://github.com/phaikawl/wade):
    * `go get -u github.com/phaikawl/wade`
* Install the demo app:
    * Client: `go get -u github.com/phaikawl/wade/testapp/ez`
    * Server: `go get -u github.com/phaikawl/wade/testapp/ez/server`

**Important**: From here, all directory paths that are referred, for example `testapp/ez`, are relative path of Wade's package directory, which is in `$GOPATH/src/github.com/phaikawl/wade`.

* Install the javascript dependencies:
    * Go to `testapp/ez/public`
    * Run `bower install`

##Running
###Fresh
**fresh** is a tool to watch for changes in the server-side Go code (located in `testapp/ez/server`), automatically compile them and reload the server.

Just go to `testapp/ez/server` and run `fresh`.

###Gopherjs
**gopherjs** compiles Go to javascript. It has a *-w* flag to watch for changes in the client-side Go code and automatically compile them. Our output target is `testapp/ez/public/app.js`.

Just go to `testapp/ez`, run

    gopherjs build -w=true -o="public/app.js"
or just run the shell script `./run_gopherjs` inside `testapp/ez` which does the same thing.

###It's up!
The demo is on `localhost:3000`, glhf!

Hopefully everything went smoothly for you. For the rest of this tour we will play with `testapp/ez` to understand how things work.

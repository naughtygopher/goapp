# Goapp

This is an opinionated guideline to structure a Go web application/service (or could be extended for any application). And my opinions formed over a span of 5+ years building web applications/services with Go.
Even though I've mentioned `go.mod` and `go.sum`. This guideline works for 1.4+ (i.e. introduction of 'internal' special directory).

P.S: This does not apply for an independent package, as their primary use is to be consumed in other applications. This is where Go's recommendation of "no unnecessary sub packages" comes into play.

In my effort to try and make things easier to understand, the structure is for an imaginary note taking web application.

```bash
|
|____internal
|    |
|    |____api
|    |    |____note.go
|    |    |____users.go
|    |
|    |____users
|    |    |____store.go
|    |    |____users.go
|    |
|    |____notes
|    |    |____notes.go
|    |
|    |____platform
|    |    |____stringutils
|    |    |____datastore
|    |
|    |____cmd
|         |____server
|              |____http
|              |    |____handlers_notes.go
|              |    |____handlers_users.go
|              |    |____http.go
|              |
|              |____grpc
|
|____docker
|    |____Dockerfile # obviously your dockerfile
|
|____lib
|    |____notes
|         |____notes.go
|
|
|____vendor
|
|____go.mod
|____go.sum
|
|____ciconfig.yml # depends on the CI/CD system you're using. e.g. .travis.yml
|____main.go
|
```

## internal

["internal" is a special directoryname in Go](https://golang.org/doc/go1.4#internalpackages), wherein any exported name/entity can only be consumed by its immediate parent.

## internal/api

The API packages is supposed to have all the APIs exposed by the application. A specific API package is created to standardize the functionality, when there are different kind of servers running. e.g. an HTTP server as well as gRPC server. In such cases, the respective "handler" functions would inturn call `api.<Method name>`. This gives a guarantee that all your APIs behave exactly the same without any accidental inconsistencies across servers. Though middleware handling is still at the cmd/server layer. e.g. access log, authentication etc. Even though this can be brought to the `api` package, it doesn't make much sense. Because middleware are mostly dependent on the server implementation. 

## internal/users

Users package is where all your actual business logic is implemented. e.g. Create a user after cleaning up the input, validation, and then put it inside a persistent data store. 

There's a `store.go` in this package which is where you write all the direct interactions with the datastore. There's an interface which is unique to the `users` package. When the file grows, and say it's more than 500 lines, you add a new one just like handlers_users.go. e.g. `store_aggregate.go` where you have numerous aggregation functionalities required for the users package.

I create a function `NewService` per package, which initializes and returns the respective package's handler. In case of users package, there's a `Users` struct. The name NewService makes sense in most cases, and just reduces the burden of thinking of a good name for such scenarios.

## internal/notes

Similar to the users package, 'notes' handles all business logic related to handling notes.

## internal/cmd

`cmd` is a shortform of "command". And its purpose is exactly that, it contains "commands" which are executed by the application. This is just an abstraction to group all the "non business logic" side of things like starting an HTTP server.

## internal/cmd/http

All HTTP related configurations and functionalities are kept inside this package. 

The naming convention followed for filenames, is also straightforward. i.e. all the HTTP handlers of a specific package/domain are grouped under `handlers_<business logic unit name>.go`. 

e.g. handlers_users.go. The advantage of naming this way is, it's easier for developers to look at and identify from a list of filenames in your file listing. e.g. on VS code it looks like this

<p align="center"><img src="https://user-images.githubusercontent.com/1092882/86511199-58ffbd80-be14-11ea-875d-0e7c37e23b95.png" alt="handlers_users.go" width="512px"/></p>


## docker

This is for anyone who is deploying their applications using Docker. Personally, I've been a fan of Docker since a few years now. I like keeping a dedicated folder for Dockerfile in anticipation of introducing multiple Docker files or other related files. e.g. [Dockerfile for Alpine as well as Debian](https://github.com/bnkamalesh/golang-dockerfile)

## lib

This name is quite explicit and if you notice, it's outside of the special 'internal' directory. So anything you define within this, is meant for consumption in external projects. 

It might seem quite redundant to add a sub-directory called 'goapp', the import path would be `github.com/bnkamalesh/goapp/lib/goapp`. Though this is not a mistake, while importing this package, you'd like to use it like this `goapp.<something>`. So if you directly put it under lib, it'd be `lib.` and that's obviously too generic and you'd have to manually setup aliases every time. Or if the package name and direcory names are different, it's a tussle with your [IDE](https://en.wikipedia.org/wiki/Integrated_development_environment).

Similarly if you have more than one package which you'd like to be made available for external consumption, you create `lib/<other>`.

## main.go

And finally the `main package`. I prefer putting the `main.go` file outside as shown here. main.go is probably going to be the ugliest package where all conventions and separation of concerns are broken. But I believe this is acceptable. The responsibility of main package is one and only one, `get things started`.


# Note

You can clone this repository any actually run the application, it'd start an HTTP server listening on port 8080 with the following routes available.

- `/` GET, the root just returns "Hello world" text response
- `/-/health` GET, returns a JSON with some basic info. I like using this path to give out the status of the app, its dependencies etc.
- `/users` POST, just exists. It doesn't do much (and will panic because there's no data store provided)

I've used [webgo](https://github.com/bnkamalesh/webgo) to setup the HTTP server (I guess I'm just biased).

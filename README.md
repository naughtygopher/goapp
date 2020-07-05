<p align="center"><img src="https://user-images.githubusercontent.com/1092882/86512217-bfd5a480-be1d-11ea-976c-a7c0ac0cd1f1.png" alt="goapp gopher" width="256px"/></p>

# Goapp

This is an opinionated guideline to structure a Go web application/service (or could be extended for any application). And my opinions formed over a span of 5+ years building web applications/services with Go. Even though I've mentioned `go.mod` and `go.sum`, this guideline works for 1.4+ (i.e. since introduction of 'internal' special directory).

P.S: This guideline is not directly applicable for an independent package, as their primary use is to be consumed in other applications. In such cases, having most or all of the package in the root is probably the best way of doing it. And that is where Go's recommendation of "no unnecessary sub packages" comes into play.

In my effort to try and make things easier to understand, the structure is explained based on an imaginary note taking web application.

```bash
|
|____internal
|    |
|    |____configs
|    |    |____configs.go
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
|    |         |____datastore.go
|    |
|    |____server
|         |____http
|         |    |____handlers_notes.go
|         |    |____handlers_users.go
|         |    |____http.go
|         |
|         |____grpc
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
|____README.md
|____main.go
|
```

## internal

["internal" is a special directoryname in Go](https://golang.org/doc/go1.4#internalpackages), wherein any exported name/entity can only be consumed by its immediate parent.

## internal/configs

Creating a dedicated configs package might seem like an overkill, but it makes a lot of things easier. In the example app provided, you see the HTTP configs are hardcoded and returned. Later you decide to change to consume from env variables. All you do is update the configs package. And further down the line, maybe you decide to introduce something like [etcd](https://github.com/etcd-io/etcd), then you define the dependency in `Configs` and update the functions accordingly. This is yet another separation of concern package, to keep the `main` package a bit less ugly.

## internal/api

The API packages is supposed to have all the APIs exposed by the application. A specific API package is created to standardize the functionality, when there are different kind of servers running. e.g. an HTTP server as well as gRPC server. In such cases, the respective "handler" functions would inturn call `api.<Method name>`. This gives a guarantee that all your APIs behave exactly the same without any accidental inconsistencies across servers. 

Though middleware handling is still at the internal/server layer. e.g. access log, authentication etc. Even though this can be brought to the `api` package, it doesn't make much sense because middleware are mostly dependent on the server implementation. 

## internal/users

Users package is where all your actual user related business logic is implemented. e.g. Create a user after cleaning up the input, validation, and then put it inside a persistent data store. 

There's a `store.go` in this package which is where you write all the direct interactions with the datastore. There's an interface which is unique to the `users` package. Such an interface is introduced to handle dependency injection as well as dependency inversion elegantly. File naming convention for store files is `store_<logical group>.go`. e.g. `store_aggregations.go`. 

`NewService` function is created in each package, which initializes and returns the respective package's handler. In case of users package, there's a `Users` struct. The name 'NewService' makes sense in most cases, and just reduces the burden of thinking of a good name for such scenarios. The Users struct here holds all the dependencies required for implementing features provided by users package.

## internal/notes

Similar to the users package, 'notes' handles all business logic related to 'notes'.

## internal/platform

Platform package contains all the packages which are to be consumed across multiple packages within the project. For instance the datastore package will be consumed by both users and notes package.

### internal/platform/datastore

The datastore package initializes `pgxpool.Pool` and returns a new instance. I'm using Postgres as the datastore in this sample app.

## internal/http

All HTTP related configurations and functionalities are kept inside this package. The naming convention followed for filenames, is also straightforward. i.e. all the HTTP handlers of a specific package/domain are grouped under `handlers_<business logic unit name>.go`. 

e.g. handlers_users.go. The advantage of naming this way is, it's easier for developers to look at and identify from a list of filenames. e.g. on VS code it looks like this

<p align="center"><img src="https://user-images.githubusercontent.com/1092882/86526182-24d8db00-beae-11ea-9681-0a31b2d67e1b.png" alt="handlers_users.go" width="512px"/></p>

### internal/http/templates

All HTML templates required for the application are to be put here. Sub directories based on the main business logic unit, e.g. users, can be created. It is highly unlikely that HTML templates used for HTTP responses are reused elsewhere in the application. Hence it justifies its location within 'server/http'.

Ideally the template is executed in HTTP handlers, and never used anywhere outside the 'server/http' package.

## docker

I've been a fan of Docker since a few years now. I like keeping a dedicated folder for Dockerfile in anticipation of introducing multiple Docker files or other files required for Docker image build.

e.g. [Dockerfile for Alpine as well as Debian](https://github.com/bnkamalesh/golang-dockerfile)

You can create the Dockerfile for the sample app provided, by:

```bash
$ git clone https://github.com/bnkamalesh/goapp.git
$ cd goapp
# Update the internal/configs/configs.go with valid datastore configuration. Or pass nil while calling user service. This would cause the app to panic when calling any API with database interaction
# Build the Docker image
$ docker build -t goapp -f docker/Dockerfile .
# and you can run the image with the following command
$ docker run -p 8080:8080 --rm -ti goapp
```

## lib

This name is quite explicit and if you notice, it's outside of the special 'internal' directory. So within this directory, is meant for consumption in external projects. 

It might seem redundant to add a sub-directory called 'goapp', the import path would be `github.com/bnkamalesh/goapp/lib/goapp`. Though this is not a mistake, while importing this package, you'd like to use it like this `goapp.<something>`. So if you directly put it under lib, it'd be `lib.` and that's obviously too generic and you'd have to manually setup aliases every time. Or if you try solving it by having the package name differ from the direcory name, it's going to be a tussle with your [IDE](https://en.wikipedia.org/wiki/Integrated_development_environment).

Another advantage is, if you have more than one package which you'd like to be made available for external consumption, you create `lib/<other>`. In this case, you reduce the dependencies which are imported to external functions. On the contrary if you put everything inside `lib` or in a single package, you'd be forcing import of all dependencies even when you'd need only a small part of it.

## vendor

I still vendor all dependencies using `go mod vendor`. vendoring is reliable and is guaranteed to not break. Chances of failure of your Go proxy for private repositories are higher compared to something going wrong with vendored packages.

## schemas

I maintain all the SQL schemas required by the project in this directory. This is not nested inside individual package because it's not consumed by the application at all. Also the fact that, actual consumers of the schema (developers, DB maintainers etc.) are varied. It's better to make it easier for all the audience rather than just developers.

## main.go

And finally the `main package`. I prefer putting the `main.go` file outside as shown here. No non-sense, straight up `go run main.go` would start the application. 'main' is probably going to be the ugliest package where all conventions and separation of concerns are broken. But I believe this is acceptable. The responsibility of main package is one and only one, `get things started`.

`cmd` directory can be added in the root for adding multiple commands. This is usually required when there are multiple modes of interacting with the application. i.e. HTTP server, CLI application etc. In which case each usecase can be initialized and started with subpackages under `cmd`. Even though Go advocates lesser use of packages, I would give higher precedence for separation of concerns at a package level.

## Integrating with ELK APM

I'm a fan of ELK APM when I first laid my eyes on it. The interation is super easy as well. In the sample app, you can check `internal/http/http.go:NewService` how APM is enabled. Once you have ELK APM setup, you need to provide the following configuration for it work.
You can [refer here](https://www.elastic.co/guide/en/apm/agent/go/current/configuration.html) for details on various configurations.

```bash
$ export ELASTIC_APM_SERVER_URL=https://apm.yourdomain.com
$ export ELASTIC_APM_SECRET_TOKEN=apmpassword
$ export ELASTIC_APM_SERVICE_NAME=goapp
$ export ELASTIC_APM_ENVIRONMENT=local
$ export ELASTIC_APM_SANITIZE_FIELD_NAMES=password,repeat_password,authorization,set-cookie,cookie
$ export ELASTIC_APM_CAPTURE_HEADERS=false
$ export ELASTIC_APM_METRICS_INTERVAL=60s
$ go run main.go
```

# Note

You can clone this repository and actually run the application, it'd start an HTTP server listening on port 8080 with the following routes available.

- `/` GET, the root just returns "Hello world" text response
- `/-/health` GET, returns a JSON with some basic info. I like using this path to give out the status of the app, its dependencies etc
- `/users` POST, to create new user
- `/users/:emailID` GET, reads a user from the database given the email id. e.g. http://localhost:8080/users/john.doe@example.com

I've used [webgo](https://github.com/bnkamalesh/webgo) to setup the HTTP server (I guess I'm just biased).

How to run?
```bash
$ git clone https://github.com/bnkamalesh/goapp.git
$ cd goapp
# Update the internal/configs/configs.go with valid datastore configuration. Or pass 'nil' while calling user service. This would cause the app to panic when calling any API with database interaction
$ go run main.go
```

## Something missing?

If you'd like to see something added, or if you feel there's something missing here. Create an issue, or if you'd like to contribute, PRs are welcome!

## Todo

- [x] Add sample Postgres implementation (for persistent store)
- [x] Add sample Redis implementation (for cache)
- [x] Add APM implementation using [ELK stack](https://www.elastic.co/apm)


## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). We all want to build reliable, resilient, maintainable applications like this adorable gopher!

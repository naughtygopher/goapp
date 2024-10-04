<p align="center"><img src="https://user-images.githubusercontent.com/1092882/86512217-bfd5a480-be1d-11ea-976c-a7c0ac0cd1f1.png" alt="goapp gopher" width="256px"/></p>

[![](https://github.com/naughtygopher/goapp/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/naughtygopher/goapp/actions)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](http://godoc.org/github.com/naughtygopher/goapp)
[![](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#tutorials)

# Goapp v1.0

This is an opinionated guideline to structure a Go web application/service (could be extended for any type of application). My opinions were formed over a span of 8+ years building web applications/services with Go, trying to implement [DDD (Domain Driven Development)](https://en.wikipedia.org/wiki/Domain-driven_design) & [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). This guideline works for 1.4+ (i.e. since introduction of the [special 'internal' directory](https://go.dev/doc/go1.4#internalpackages)).

P.S: This guideline is not directly applicable for an independent package, as their primary use is to be consumed in other applications. In such cases, having most or all of the package code in the root is probably the best way of doing it.

The structure is explained based on a note taking web application (with hardly any features implemented 🤭).

## Table of contents

1. [Directory structure](#directory-structure)
2. [Configs package](#internalconfigs)
3. [API package](#internalapi)
4. [Users](#internalusers) (would be common for all such business logic / domain units, 'usernotes' being similar to users) package.
5. [Testing](#internalusers_test)
6. [pkg package](#internalpkg)
   - 6.1. [datastore](#internalpkgdatastore)
   - 6.2. [logger](#internalpkglogger)
7. [HTTP server](#internalhttp)
   - 7.1. [templates](#internalhttptemplates)
8. [lib](#lib)
9. [vendor](#vendor)
10. [docker](#docker)
11. [schemas](#schemas)
12. [main.go](#maingo)
13. [Error handling](#error-handling)
14. [Dependency flow](#dependency-flow)
15. [Integrating with ELK APM](#integrating-with-elk-apm)
16. [Note](#note)

## Directory structure

```bash
├── cmd
│   ├── server
│   │   ├── grpc
│   │   │   └── grpc.go
│   │   └── http
│   │       ├── handlers.go
│   │       ├── handlers_usernotes.go
│   │       ├── handlers_users.go
│   │       ├── http.go
│   │       └── web
│   │           └── templates
│   │               └── index.html
│   └── subscribers
│       └── kafka
│           └── kafka.go
├── docker
│   ├── docker-compose.yml
│   └── Dockerfile
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   ├── api.go
│   │   ├── usernotes.go
│   │   └── users.go
│   ├── configs
│   │   └── configs.go
│   ├── pkg
│   │   ├── apm
│   │   │   ├── apm.go
│   │   │   ├── grpc.go
│   │   │   ├── http.go
│   │   │   ├── meter.go
│   │   │   ├── prometheus.go
│   │   │   └── tracer.go
│   │   ├── logger
│   │   │   ├── default.go
│   │   │   └── logger.go
│   │   ├── postgres
│   │   │   └── postgres.go
│   │   └── sysignals
│   │       └── sysignals.go
│   ├── usernotes
│   │   ├── store_postgres.go
│   │   └── usernotes.go
│   └── users
│       ├── store_postgres.go
│       └── users.go
├── lib
│   └── goapp
│       ├── goapp.go
│       ├── go.mod
│       └── go.sum
├── LICENSE
├── main.go
├── README.md
└── schemas
    ├── functions.sql
    ├── user_notes.sql
    └── users.sql
```

## internal

["internal" is a special directory name in Go](https://go.dev/doc/go1.4#internalpackages), wherein any exported name/entity can only be consumed within its immediate parent or any other packages within internal directory.

## internal/configs

Creating a dedicated configs package might seem like an overkill, but it makes things easier. In the app, you see the HTTP configs are hardcoded and returned. Later you decide to change to consume from env variables. All you do is update the configs package. And further down the line, maybe you decide to introduce something like [etcd](https://github.com/etcd-io/etcd), then you define the dependency in `Configs` and update the functions accordingly. This is yet another separation of concern package, to try and keep `main` tidy.

## internal/api

The API package is supposed to have all the APIs _*exposed*_ by the application. A dedicated API package is created to standardize the functionality, when there are different kinds of services running. e.g. an HTTP & a gRPC server, a Kafka & Pubsub subscriber etc. In such cases, the respective "handler" functions would inturn call `api.<Method name>`. This gives a guarantee that all your APIs behave exactly the same without any accidental inconsistencies across different I/O methods. It also helps consolidate which functionalities are expcted to be exposed outside of the application via API. There could be a variety of exported functions in the domain packages, which are not meant to communicate with anything outside the application rather to be used among other domain packages.

But remember, middleware handling is still at the internal/server layer. e.g. access log, authentication etc. Even though this can be brought to the `api` package, it doesn't make much sense because middleware are mostly dependent on the server/handler implementation. e.g. HTTP method, path etc.

## internal/users

Users package is where all your actual user related _business logic_ is implemented. e.g. Create a user after cleaning up the input, validation, and then store it inside a persistent datastore.

The `store_postgres.go` in this package is where you write all the direct interactions with the datastore. There's an interface which is unique to the `users` package. It is used to handle dependency injection as well as dependency inversion elegantly. The file naming convention I follow is to have the word `store` in the beggining, suffixed with `_<db name>`. Though I think it's ok name it based on a logical group, e.g. `store_registration`, `store_login` etc.

`NewService/New` function is created in each package, which initializes and returns the respective package's feature _implementor_. In case of users package, it's the `Users` struct. The name 'NewService' makes sense in most cases, and just reduces the burden of thinking of a good name for such scenarios. The Users struct here holds all the dependencies required for implementing features provided by users package.

## internal/users_test

There's quite a lot of discussions about achieveing and maintaining 100% test coverage or not. 100% coverage sounds very nice, but might not always be practical or at times not even possible. What I like doing is, writing unit test for your core business logic, in this case 'Sanitize', 'Validate' etc are my business logic.

It is important for us to understand the purpose of unit tests. The sole purpose of unit test is unironically "test the purpose of the unit/function". It is _*not*_ to check the implementation, how it's done, how much time it took, how efficient it is etc. The sole purpose is to validate "what it does". This is why you see a lot of unit tests will have hardcoded values, because those are reliable/verified human input which we validate against.

Once you develop the habit of writing unit tests for [pure functions](https://en.wikipedia.org/wiki/Pure_function) and get the hang of it. You automatically start breaking down big functions into smaller _*testable*_ functions/units (this is the best outcome, and what we'd love to have). When you _layer_ your application, datastore is ideally just a utility (_implementation detail_ in [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) parlance), and if you can implement your business logic with pure functions alone, not dependent on such utlities, that'd be perfect! Though in most cases you'd have dependencies like database, queue, cache etc. But to keep things as _pure_ as possible, we bridge the gap using Go interfaces. Refer to `store.go`, the business logic functions are oblivious to the underlying technology (RDBMS, NoSQL, CSV etc.).

Always writing the entire business logic within the app is not necessary or sometimes extremely difficult, rather make use of features provided by databases and other tools. e.g. Database can do joins, sort etc. Though when using such features, it's best that the function signature hints at this. e.g. `GetUserNotes(ctx, userID) []Note` is a name which hints at the joining of User and Note. This way, if we decide to switch database which does not support join, we still know the expected behaviour from the data store function.

### integration tests

In case of writing integration tests, i.e. when you make API calls from outside the app to test functionality, I prefer using actual running instances of dependencies instead of mocks. Especially in case of databases, or any such easy to use dependency. Though if the dependency is an external service's APIs, mocks are probably the best available option.

## internal/usernotes

Similar to the users package, 'usernotes' handles all business logic related to user's notes.

## internal/pkg

pkg package contains all the packages which are to be consumed across multiple packages within the project. For instance the _*postgres*_ package will be consumed by both users and usernotes package.

### internal/pkg/postgres

The postgres package initializes `pgxpool.Pool` and returns a new instance. Though a seemingly redundant package only for initialization, it's useful to do all the default configuration which we want standardized across the application. An example is to wrap the driver, or functions for [APM](https://en.wikipedia.org/wiki/Application_performance_management). The screenshots below show how APM can help us monitor our application.

<p align="center">
<img src="https://user-images.githubusercontent.com/1092882/86710556-baa07180-c038-11ea-8924-3b4d61db1476.png" alt="APM overall" width="384px" height="256px" style="margin-right: 16px" />
<img src="https://user-images.githubusercontent.com/1092882/86710547-b83e1780-c038-11ea-9829-b5585b3d599b.png" alt="APM 1 API" width="384px" height="256px" />
</p>

### internal/pkg/logger

I usually define the logging interface as well as the package, in a private repository (internal to your company e.g. vcs.yourcompany.io/gopkgs/logger), and is used across all services. Logging interface helps you to easily switch between different logging libraries, as all your apps would be using the interface **you** defined (interface segregation principle from SOLID). But here I'm making it part of the application itself as it has fewer chances of going wrong when trying to cater to a larger audience.

**Logging might sound trivial but there are a few questions around it:**

1. Should it be made a dependency of all packages, or can it be global?

Logging just like any other dependency, is a dependency. And in most cases it's better to write packages (code in general) which have as few dependencies as practically possible. This is a general principle, fewer dependencies make a lot of things easier like maintainability, testing, porting, refactoring, etc. And creating singleton Globals bring in restrictions, also it's a dependency nevertheless. Global instances have another issue, it doesn't give you flexibility when you need varying functionality across different packages (since it's global, it's common for all consumers). E.g. in one package you'd like to have debug logs, and in the other you'd only want errors. So in my opinion, it's better not to use a global instance, but have global functions which implement the default behaviour for all your packages which do not have any custom requirements.

2. Where would you do it? Should you bubble up errors and log at the parent level, or write where the error occurs?

Keeping it at the root/outermost layer helps make things easier because you need to worry about injecting logging dependency only in this package. And easier to control it in general. i.e. One less thing to worry about in majority of the code.

For developers, while troubleshooting (which is one of the foremost need for logging), the line number along with filename helps a lot. Then it's obvious, log where the error occurs, right?

Over the course of time, I found it's not really obvious. The more nested function calls you have, higher the chances of redundant logging. And setting up guidelines to only log at the origin of error is also not easy. It's easy to get confused which level should be considered the origin (especially when there's deep nesting fn1 -> fn2 -> fn3 -> fn4). Thus I prefer logging at the Handlers layer, [with annotated errors](https://pkg.go.dev/errors)(using the '%w' verb in `fmt.Errorf`) to trace its origin. Recently I introduced a [minimal error handling package](https://github.com/bnkamalesh/errors/) which gives long file path, line number of the origin of error, stacktrace etc. as well as help set user friendly messages for API response. Now all the HTTP handlers return an error, and there's a wrapper to handle the logging as well as responding to the HTTP request.

There are some exceptions to logging at the outer most layer. In case of async functions, where the caller function is doing _fire and forget_, it's still important for us to be able to troubleshoot issues within the async function. Another scenario where it'd be important to log error immediately would be; read-through cache, where the app is expected to simply read info from the primary database if the cache is a miss or even if the cache DB is down. In such cases, the API would successfully respond, and for us to find out the cache DB is down, we'd have to rely on logs.

## cmd/server/http

All HTTP related configurations and functionalities are kept inside this package. The naming convention followed for filenames, is also straightforward. i.e. all the HTTP handlers of a specific package/domain are grouped under `handlers_<business logic unit name>.go`. The special mention of naming handlers is because, often for decently large web applications (especially when building REST-ful services) you end up with a lot of handlers. I have services with 100+ handlers for individual APIs, so keeping them organized helps.

e.g. handlers_users.go. The advantage of naming this way is, it's easier for developers to look at and identify from a list of filenames. e.g. on VS code it looks like this, even if you list the files from a basic shell, it'd be sorted/grouped.

<p align="center"><img src="https://user-images.githubusercontent.com/1092882/86526182-24d8db00-beae-11ea-9681-0a31b2d67e1b.png" alt="handlers_users.go" width="512px"/></p>

### internal/server/http/web/templates

All HTML templates required for the application are to be put here. Sub directories based on the main business logic unit, e.g. we/templates/users, can be created if required. It is highly unlikely that HTML templates used for HTTP responses are reused elsewhere in the application. Hence it justifies its location within 'server/http'. Other static files shall also be made part of the `web` directory like `web/static/images`, `web/static/js` etc. Feel free to [embed](https://pkg.go.dev/embed) templates, static files etc.

## lib

This name is quite explicit and if you notice, it's outside of the special 'internal' directory. So any exported name or entity within this directory, is meant to be used in external projects.

It might seem redundant to add a sub-directory called 'goapp', the import path would be `github.com/naughtygopher/goapp/lib/goapp`. Though this is not a mistake, while importing this package, you'd use it as follows `goapp.<something>`. Rather if you directly put it under lib, it'd be `lib.<something>` and that's obviously too generic and you'd have to manually setup aliases every time. Or if you try solving it by having the package name which differ from the direcory name, it's going to be a tussle with your [IDE](https://en.wikipedia.org/wiki/Integrated_development_environment).

Another advantage is, if you have more than one package which you'd like to be made available for external consumption, you create `lib/<other>`. In this case, you reduce the dependencies which are imported to external functions. On the contrary if you put everything inside `lib` or in a single package, you'd be forcing to import of all dependencies even when you'd need only a small part of it.

## vendor (deprecated)

I've stopped vendoring packages, and have been relying on downloading packages on every build (when no cache). It hasn't failed me for the past few years I've been using it.

## docker

I've been a fan of Docker since a few years now (~2016). I like keeping a dedicated folder for Dockerfile, in anticipation of introducing multiple Docker files or maintaining other files required for Docker image build.

e.g. [Dockerfiles for Go applications](https://github.com/bnkamalesh/golang-dockerfile)

You can create the Docker image for the sample app provided:

```bash
$ git clone https://github.com/naughtygopher/goapp.git
$ cd goapp
# Update the internal/configs/configs.go with valid datastore configuration. Or pass nil while calling user service. This would cause the app to panic when calling any API with database interaction
# Build the Docker image
$ docker build -t goapp -f docker/Dockerfile .
# and you can run the image with the following command
$ docker run -p 8080:8080 --rm -ti goapp
```

## schemas

All the SQL schemas required by the project in this directory. This is not nested inside individual package because it's not consumed by the application at all. Also the fact that, actual consumers of the schema (developers, DB maintainers etc.) are varied. It's better to make it easier for all the audience rather than just developers. Even if you use NoSQL databases, your application would need some sort of schema to function, which can still be maintained inside this.

I've recently started using [sqlc](https://sqlc.dev/) for code generation for all SQL interactions (and love it!). I use [Squirrel](https://github.com/Masterminds/squirrel) whenever I need to dynamically build queries. E.g. when updating a table, you want to update only certain columns based on the input.

Even migrations can be maintained in a directory in the root, but it's best to keep the application never be responsible for database setup. i.e. let migrations, index creation etc. be handled outside the scope of the application itself. For instance, it's very easy to create deadlocks with databases if it's part of the application, when you deploy the application in a _horizontally_ scaled model. Though there is nothing wrong in keeping the migration files within the same repository. Below are a few tools to use for migration

1. [Golang Migrate](https://github.com/golang-migrate/migrate)
2. [goose](https://github.com/golang-migrate/migrate)

## main.go

Finally the `main package`. I prefer putting the `main.go` file outside as shown here. No non-sense, straight up `go run main.go` would start the application (provided the required configurations are available). 'main' is probably going to be the ugliest package where all conventions and separation of concerns are broken, but this is acceptable. The responsibility of main package is one and only one, **get things started**.

`cmd` directory can be added in the root for adding multiple commands. This is usually required _when there are multiple modes of interacting with the application_. i.e. HTTP server, CLI etc. In which case each usecase can be initialized and started with subpackages under `cmd`. Even though Go advocates fewer use of packages, I would give higher precedence for separation of concerns at a package level to keep things tidy. And even the main.go can be in `cmd/main.go`.

## Error handling

After years of trying different approaches, I finally caved and a created custom [error handling package](https://github.com/bnkamalesh/errors) to make troubleshooting and responding to APIs easier, p.s: it's a drop-in replacement for Go builtin errors. More often than not, we log full details of errors and then respond to the API with a cleaner/friendly message. If you end-up using the [errors](https://github.com/bnkamalesh/errors) package, there's only one thing to follow. Any error returned by an external (external to the project/repository) should be wrapped using the respective helper method. e.g. `errors.InternalErr(err, "<user friendly message>")` where err is the original error returned by the external package. If not using the custom error package, then you would have to annotate all the errors with relevant context info. e.g. `fmt.Errorf("<more info> %w", err)` throughout the calling chain to get a stacktrace. If you're annotating errors all the way, the user response has still to be handled separately. In which case, HTTP status code and the custom messages are better handled in the handler layer.

## Dependency flow

<p align="center">
<img src="https://user-images.githubusercontent.com/1092882/104085767-f5999100-5277-11eb-808a-5fd9b6776ad6.png" alt="Dependency flow between the layers" width="768px"/>
</p>

## Integrating Open telemetry for instrumentation

[Open telemetry](https://opentelemetry.io/) released their [first stable version,v1.23.0, in Feb 2024](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.23.0), and is supported by most APM/instrumentation providers.

You can find [Go's Open telemetry libraries here](https://opentelemetry.io/docs/instrumentation/go/). I have added sample for usage for HTTP server and gRPC in this repository.

# Note

You can clone this repository and try running the application, it'd start an HTTP server listening on port 8080 with the following routes available.

- `/` GET, the root just returns "Hello world" text response
- `/-/health` GET, returns a JSON with some basic info. I like using this path to give out the status of the app, its dependencies etc
- `/users` POST, to create new user
- `/users/:emailID` GET, reads a user from the database given the email id. e.g. http://localhost:8080/users/john.doe@example.com

I've used [webgo](https://github.com/bnkamalesh/webgo) to setup the HTTP server (I guess I'm biased ¯\\ (ツ) /¯ ). Though there's no compulsion that you do the same, you can pick a framework of your choice! Though stick to the framework's structure if they have any recommendations. Otherwise, goapp is the way to _go_, yay!

How to run?

```bash
$ git clone https://github.com/naughtygopher/goapp.git
$ cd goapp
# Update the internal/configs/configs.go with valid datastore configuration. Or pass 'nil' while calling user service. The app wouldn't start if no valid configuration is provided.
$ TEMPLATES_BASEPATH=${PWD}/cmd/server/http/web/templates go run main.go | sed 's/\\n/\n/g;s/\\t/\t/g'
```

## Use Go app to start a new project

[gonew](https://go.dev/blog/gonew) lets you download a new Go module, and name it with a custom Go module name.

```bash
$ gonew github.com/naughtygopher/goapp@latest my.app
$ cd my.app
```

## Something missing?

If you'd like to see something added, or if you feel there's something missing here. Create an issue, or if you'd like to contribute, PRs are welcome!

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). We all want to build reliable, resilient, maintainable applications like this adorable gopher!

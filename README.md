# ![RealWorld Example App](logo.png)

> ### [Go net/http](//golang.org) codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.


### [Demo](https://demo.realworld.io/)&nbsp;&nbsp;&nbsp;&nbsp;[RealWorld](https://github.com/gothinkster/realworld)


This codebase was created to demonstrate a fully fledged fullstack application built with **Go net/http library** including CRUD operations, authentication, routing, pagination, and more.

We've gone to great lengths to adhere to the **Golang** community styleguides & best practices.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.


# How it works

The project structure was inspired by two posts on [Ben Johnson's](https://twitter.com/benbjohnson) blog which can be found [here](https://www.gobeyond.dev/packages-as-layers/) and [here](https://www.gobeyond.dev/standard-package-layout/).

# Getting started
This project uses Go version 1.17 and postgresql 14
You also need to have [migrate](https://github.com/golang-migrate/migrate) tool installed to run all migrations against the database
#### Locally
- make sure [Go](https://golang.org/dl) is installed on your machine.
- make sure to have the postgresql database installed locally or remote
- set the .env file or env var as shown in the .env.example file
- run the migrations in postgres/migrations or run `make run-migration`
- fetch all dependencies using `go mod download`
- run `make run` to start the server locally
 

#### TODO
- Revisit error handling

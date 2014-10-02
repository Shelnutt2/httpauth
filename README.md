# Go Session Authentication
[![Build Status](https://travis-ci.org/apexskier/httpauth.svg?branch=master)](https://travis-ci.org/apexskier/httpauth)
[![GoDoc](https://godoc.org/github.com/apexskier/httpauth?status.png)](https://godoc.org/github.com/apexskier/httpauth)

This package uses the [Gorilla web toolkit](http://www.gorillatoolkit.org/)'s
sessions package to implement a user authentication and authorization system
for Go web servers.

Multiple user data storage backends are available, and new ones can be
implemented relatively easily.

- [File based](https://godoc.org/github.com/apexskier/goauth#NewGobFileAuthBackend) ([gob](http://golang.org/pkg/encoding/gob/))
- [Various SQL Databases](https://godoc.org/github.com/apexskier/httpauth#NewSqlAuthBackend)
- [MongoDB](https://godoc.org/github.com/apexskier/httpauth#NewMongodbBackend)

Access can be restricted by a users' role.

Uses [bcrypt](http://codahale.com/how-to-safely-store-a-password/) for password
hashing.

Run `go run server.go` from the examples directory and visit `localhost:8009`
for an example. You can login with the username and password "admin".

Tests can be run by simulating Travis CI's build environment. A mysql database
must be running on the default port with a user "travis" with no password and a
database "httpauth_test". A mongodb database must be running with open access
as well. There's a very unsafe script --- `start-test-env.sh` that will do this
for you.

**Note**

This is the first time I've worked with implementing the details of cookie
storage, authentication or any sort of real security. There are no guarantees
that this will work as expected, but I'd love feedback. If you have any issues
or suggestions, please [let me
know](https://github.com/Wombats/goauth/issues/new).

### TODO

- User roles - modification
- SMTP email validation (key based)
- More backends

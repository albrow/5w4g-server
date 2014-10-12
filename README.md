5W4G Server
===========

The server-side code for 5W4G.com

Prerequisites
-------------

- Go version >= 1.3
	- [Installation instructions](http://golang.org/doc/install)
	- [Post-install instructions](http://golang.org/doc/code.html)
- Redis version >= 2.8.9
	- [Installation instructions](http://redis.io/download)
	- [Post-install instructions](http://redis.io/topics/quickstart)
- [Git](http://git-scm.com/downloads)
- [Bazaar](http://bazaar.canonical.com/en/)
	- If you're on Mac OS and you have homebrew, just run `brew install bazaar`
- [Mercurial](http://mercurial.selenic.com/)
	- If you're on Mac OS and you have homebrew, just run `brew install mercurial`

Getting Up and Running
----------------------

Install dependencies:

```bash
go get ./...
```

Start redis. Depending on your setup, you may be able to just run `redis-server`

Run the server:

```bash
go run server.go
```

If you want, you can install a tool called [fresh](https://github.com/pilu/fresh),
which will automatically restart the application when you make changes to the source
code.

To run the server with fresh, just use:

```bash
fresh
```

Runtime Environments
--------------------

The 5w4g server has 3 different runtime environments, each of which uses a different database and runs on a different port.
The environments are configured in config/config.go. You can change the runtime environment with the `GO_ENV` variable. So, for
example to run in the test environment, run `GO_ENV=test go run server.go` or `GO_ENV=test fresh`

#### Development
The default environment if none is specified. Used for development on a local machine. Not for use on a
remote server. Prints out full messages and a line number if there is an internal server error.

#### Production
For use on a remote server. Does not print out full error messages or a line number when there is an internal
server error and instead prints out a generic message. This is for security reasons.

#### Test
Used for running tests, i.e. all the code in the tests folder will attempt to connect to a server
running in the test environment. When running in the test environment, the database is erased everytime
you restart the server.
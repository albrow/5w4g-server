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

Error Codes
-----------

The following response codes will be returned by the server:

**200: OK**  
The request was successful.

**401: Unauthorized**  
The request cannot be processed without authentication. In this case the user
should be redirected to the sign in page.

**403: Forbidden**  
The user is trying to send a request which he/she is not authorized to send.
An example would be a non-admin user trying to create or remove items.

**422: Unprocessable Entity**  
The paramaters for the request were either incorrect or improperly formatted.
You should check the errors field to find out what went wrong and display the
error(s) to the user.

**500: Internal Server Error**  
An error occured on the server-server side. Will always return an errors field
in the JSON response which contains details about the error. If the runtime environment
is set to production, all internal server errors will simply contain the text: "Sorry
there was a problem." for security reasons.
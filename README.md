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

### Just run the server

If you would like to simply run the server in dev mode and don't intend to make any changes,
after you have installed the prerequisites above, you can use two commands to install the
server, compile and run it.

First, get and compile the server code with

```bash
go get github.com/albrow/5w4g-server
```

Then start redis. Depending on your setup, you may be able to just run `redis-server`

Finally, execute the binary with

```
5w4g-server
```

These instructions assume you have followed the go [post-install instructions](http://golang.org/doc/code.html).
Most importantly, you will need to have GOPATH set and add GOPATH/bin to your PATH.


### Modify the source code and/or run the tests

If you plan to modify the source code or run the tests, you will need to follow slightly different
instructions. First, clone this repository, then change into the root directory for the source code:
`cd $GOPATH/src/github.com/albrow/5w4g-server`.

Install all the go dependencies:

```bash
go get ./...
```

Start redis. Depending on your setup, you may be able to just run `redis-server`

Run the server:

```bash
go run server.go
```

It is highly recommended that you install [fresh](https://github.com/pilu/fresh),
which will automatically restart the application when you make changes to the source
code.

To run the server with fresh, just use:

```bash
fresh
```

### Running the tests

Assuming you have already cloned the repo and installed the dependencies,
you must first start a server running in the test environment:

```bash
GO_ENV=test fresh
```

Then, with the server still running in the test environment, run the tests in a separate
process:

```
go test ./...
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


REST Endpoints
--------------

#### POST /admin/sessions
Purpose: Create a new session (i.e., sign in) as an admin user

URL Parameters: none

Body Parameters:
(fields with an asterisk are required)

| Field    			 | Description     |
| ---------------- | --------------- |
| email\*          | The admin user's email address. Must be properly formatted. |
| password\*       | The admin user's password. |

Response:

| Field    		   | Type      | Description     |
| --------------- | --------- | --------------- |
| admin           | object    | The admin user. Contains fields such as email and id. |
| errors          | object    | The errors that occured (if any). |
| message         | string    | A message from the server (if any). |
| alreadySignedIn | boolean   | Whether or not the user was already signed in when the request was sent. | 

Example Responses:

```json
{
   "admin": {
      "email": "admin@5w4g.com",
      "id": "2AmRlXcIDvmc8tXVndd09p"
   },
   "alreadySignedIn": true,
   "message": "You were already signed in!"
}
```

```json
 {
    "errors": {
        "email": [
            "email is required.",
            "email must be correctly formatted."
        ],
        "password": [
            "password is required."
        ]
    }
}
```

#### DELETE /admin/sessions
Purpose: Delete an existing session (i.e., sign out)

URL Parameters: none

Body Parameters: none

Response:

| Field    		    | Type      | Description     |
| ---------------- | --------- | --------------- |
| errors           | object    | The errors that occured (if any). |
| message          | string    | A message from the server (if any). |
| alreadySignedOut | boolean   | Whether or not the user was already signed out when the request was sent. | 

Example Response:

```json
{
   "alreadySignedOut": false,
   "message": "You have been signed out."
}
```

#### GET /admin/sessions
Purpose: Get the user data corresponding to the current session (i.e., sign out)

URL Parameters: none

Body Parameters: none

Response:

| Field    		    | Type      | Description     |
| ---------------- | --------- | --------------- |
| admin            | object    | The admin user. Returned iff signedIn is true. Contains fields such as email and id. |
| errors           | object    | The errors that occured (if any). |
| message          | string    | A message from the server (if any). |
| signedIn         | boolean   | Whether or not the user was signed in. | 

Example Responses:

```json
{
   "admin": {
      "email": "admin@5w4g.com",
      "id": "2AmRlXcIDvmc8tXVndd09p"
   },
   "message": "You are signed in.",
   "signedIn": true
}
```

```json
{
    "message": "You are not signed in.",
    "signedIn": false
}
```

#### POST /admin/users
**Requires Admin Authentication**

Purpose: Create a new admin user

URL Parameters: none

Body Parameters:
(fields with an asterisk are required)

| Field             | Description     |
| ----------------- | --------------- |
| email\*           | The admin user's email address. Must be properly formatted. |
| password\*        | The admin user's password. Must be at least 8 characters long. |
| confirmPassword\* | The admin user's password again. Must match password. |

Response:

| Field            | Type      | Description     |
| ---------------- | --------- | --------------- |
| admin            | object    | The admin user. Contains fields such as email and id. |
| errors           | object    | The errors that occured (if any). |
| message          | string    | A message from the server (if any). |


Example Responses:

```json
{
    "admin": {
        "email": "new@example.com",
        "id": "DnlK3zdiqsv6Hwzdnddajl"
    },
    "message": "New admin user created!"
}
```

```json
{
    "errors": {
        "email": [
            "that email address is already taken."
        ],
        "password": [
            "password must be at least 8 characters long."
        ]
    }
}
```

#### GET /admin/users
**Requires Admin Authentication**

Purpose: List all existing admin users

URL Parameters: none

Body Parameters: none

Response:

| Field             | Type      | Description     |
| ----------------- | --------- | --------------- |
| admins            | object    | A javascript array of admin users. Each contains fields such as email and id. |
| errors            | object    | The errors that occured (if any). |


Example Responses:

```json
{
    "admins": [
        {
            "email": "admin@5w4g.com",
            "id": "2AmRlXcIDvmc8tXVndd09p"
        },
        {
            "email": "a@b.c",
            "id": "DnlK3zdiqsv6Hwzdnddajl"
        },
        {
            "email": "new@example.com",
            "id": "OQyxUYZU2Cd2pgRFnddabo"
        }
    ]
}
```

#### DELETE /admin/users/:id
**Requires Admin Authentication**

Purpose: Delete an existing admin user

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the admin user you want to delete |

Body Parameters: none

Response:

| Field             | Type      | Description     |
| ----------------- | --------- | --------------- |
| admins            | object    | A javascript array of admin users. Each contains fields such as email and id. |
| errors            | object    | The errors that occured (if any). |


Example Responses:

```json
{
    "admins": [
        {
            "email": "admin@5w4g.com",
            "id": "2AmRlXcIDvmc8tXVndd09p"
        },
        {
            "email": "a@b.c",
            "id": "DnlK3zdiqsv6Hwzdnddajl"
        },
        {
            "email": "new@example.com",
            "id": "OQyxUYZU2Cd2pgRFnddabo"
        }
    ]
}
```

```json
{
    "errors": {
        "error": [
            "You can't delete yourself, bro!"
        ]
    }
}
```

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
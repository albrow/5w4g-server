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

### Generating RSA Keys

5w4g-server uses JSON Web Tokens for authentication. To sign the tokens, you will need to
generate id.rsa and id.rsa.pub files in the config directory. Do this with the following
commands on a unix-type system:

``` bash
openssl genrsa -out config/id.rsa 1024 
openssl rsa -in config/id.rsa -pubout > config/id.rsa.pub 
```


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

Authentication
--------------

5w4g-server uses
[JSON Web Tokens](https://developer.atlassian.com/static/connect/docs/concepts/understanding-jwt.html)
(JWTs) for authentication. When you sign in, the server will respond with a token which
the client is responsible for storing. For authenticated requests, you should include
an `Authorization` header, the value of which should be the word "Bearer" followed by a space
and then the full JWT. For example:

```
GET /resource HTTP/1.1
Host: server.example.com
Authorization: Bearer mF_9.B5f-4.1JqM
```

The tokens issued by 5w4g-server include the following claims:

| Claim            | Description     |
| ---------------- | --------------- |
| adminId          | A unique identifier for an admin user |
| exp              | The expiration date of the token, as UTC unix time |
| ita              | The time the token was originally issued, as UTC unix time |

The claims are unencrypted, but protected from modification by a signature. Clients can
read a stored token to determine the adminId and whether or not it is expired.

Runtime Environments
--------------------

The 5w4g server has 3 different runtime environments, each of which uses a different
database and runs on a different port. The environments are configured in config/config.go.
You can change the runtime environment with the `GO_ENV` variable. So, for
example to run in the test environment, run `GO_ENV=test go run server.go` or
`GO_ENV=test fresh`

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

#### POST /admin/sign_in
Purpose: Sign in an admin user (i.e. get a fresh, valid token)

URL Parameters: none

Body Parameters:
(fields with an asterisk are required)

| Field            | Description     |
| ---------------- | --------------- |
| email\*          | The admin user's email address. Must be properly formatted. |
| password\*       | The admin user's password. |

Response:

| Field           | Type      | Description     |
| --------------- | --------- | --------------- |
| token           | string    | A JSON Web Token which can be used for authentication of future requests. |
| errors          | array     | The errors that occured (if any). |



Example Responses:

```json
{
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbklkIjoiVGFXY25NOXIyN1h4OVpnQm5kZGJjMiIsImV4cCI6MTQyMTY5ODQyMSwiaWF0IjoxNDE5MTA2NDIxfQ.0_GtGwP3XGwcFIYnF2EcKNUbl3bRKRgYvWCCF89uxes"
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
| errors           | array     | The errors that occured (if any). |


Example Responses:

```json
{
    "admin": {
        "email": "new@example.com",
        "id": "DnlK3zdiqsv6Hwzdnddajl"
    }
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
| admins            | array     | A javascript array of admin users. Each contains fields such as email and id. |
| errors            | array     | The errors that occured (if any). |


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
| errors            | array     | The errors that occured (if any). |

Note: if the request was successful, the response will simply be an empty JSON object.

Example Responses:

```json
{}
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

#### POST /admin/items
**Requires Admin Authentication**

Purpose: Create a new item

URL Parameters: none

Body Parameters:
(fields with an asterisk are required)

| Field            | Description     |
| ---------------- | --------------- |
| name\*           | The name of the item. Must be unique. |
| description\*    | The description for the item. Should be a sentence or two. |
| price\*          | The price of the item in dollars (decimal points allowed). |
| imageUrl\*       | The url of the image for the item (e.g. hosted on dropbox or s3/cloudfront). |

Response:

| Field            | Type      | Description     |
| ---------------- | --------- | --------------- |
| item             | object    | The item object. |
| errors           | array     | The errors that occured (if any). |


Example Responses:

```json
{
    "item": [
        {
            "name": "Ice Cube Sticker",
            "imageUrl": "http://placehold.it/350x350",
            "price": 3,
            "description": "This sticker is really cool. Ice cold, actually.",
            "id": "k4FbclRdLYXVfELnndelad"
        }
    ]
}
```

```json
{
    "errors": {
        "imageUrl": [
            "imageUrl is required."
        ],
        "name": [
            "that item name is already taken."
        ],
        "price": [
            "price must be greater than 0.000000."
        ]
    }
}
```

#### GET /admin/items/:id
**Requires Admin Authentication**

Purpose: Get a single existing items

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the item you want to get |

Body Parameters: none

Response:

| Field             | Type      | Description      |
| ----------------- | --------- | ---------------- |
| item              | object    | The item object. |
| errors            | array     | The errors that occured (if any). |

Note: if the request was successful, the response will simply be an empty JSON object.

Example Responses:

```json
{
    "item": {
        "name": "Nothing",
        "imageUrl": "http://placehold.it/350x350",
        "price": 10000000000000000,
        "description": "Pay us money for nothing. There is no sticker and no item of any kind.",
        "id": "9RlIVFnDQ5IHCwUpndemxm"
    }
}
```

```json
{
    "errors": {
        "error": [
           "dial tcp 127.0.0.1:6379: connection refused"
        ]
    }
}
```

#### GET /admin/items
**Requires Admin Authentication**

Purpose: List all existing items

URL Parameters: none

Body Parameters: none

Response:

| Field      | Type      | Description     |
| ---------- | --------- | --------------- |
| items      | array     | A javascript array of items. |
| errors     | array     | The errors that occured (if any). |


Example Responses:

```json
{
    "items": [
        {
            "name": "Sticker Sticker",
            "imageUrl": "http://placehold.it/350x350",
            "price": 42,
            "description": "Yo dawg, I herd you like stickers, so I put a sticker on your sticker so you can stick a sticker sticker to your stickers. ",
            "id": "2MxIL1sHBvXZKUMDndelnw"
        },
        {
            "name": "Ice Cube Sticker",
            "imageUrl": "http://placehold.it/350x350",
            "price": 3,
            "description": "This sticker is really cool. Ice cold, actually.",
            "id": "k4FbclRdLYXVfELnndelad"
        }
    ]
}
```


#### DELETE /admin/items/:id
**Requires Admin Authentication**

Purpose: Delete an existing item

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the item you want to delete |

Body Parameters: none

Response:

| Field             | Type      | Description     |
| ----------------- | --------- | --------------- |
| errors            | array     | The errors that occured (if any). |

Note: if the request was successful, the response will simply be an empty JSON object.

Example Responses:

```json
{}
```

```json
{
    "errors": {
        "error": [
           "dial tcp 127.0.0.1:6379: connection refused"
        ]
    }
}
```

#### PUT /admin/items/:id
**Requires Admin Authentication**

Purpose: Update an existing item

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the item you want to update |

Body Parameters:
(fields with an asterisk are required)

| Field         | Description     |
| ------------- | --------------- |
| name          | The name of the item. Must be unique. |
| description   | The description for the item. Should be a sentence or two. |
| price         | The price of the item in dollars (decimal points allowed). |
| imageUrl      | The url of the image for the item (e.g. hosted on dropbox or s3/cloudfront). |

Response:

| Field            | Type      | Description     |
| ---------------- | --------- | --------------- |
| item             | object    | The item object. |
| errors           | array     | The errors that occured (if any). |


Example Responses:

```json
{
    "item": {
        "name": "This name was updated",
        "imageUrl": "http://placehold.it/350x350",
        "price": 9000,
        "description": "This description was also just updated",
        "id": "k4FbclRdLYXVfELnndelad"
    }
}
```

```json
{
    "errors": {
        "imageUrl": [
            "imageUrl is required."
        ],
        "name": [
            "that item name is already taken."
        ],
        "price": [
            "price must be greater than 0.000000."
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
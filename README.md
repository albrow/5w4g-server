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

### Add the necessary environment variables

As of now, 5w4g-server requires the following environment variables

| Key                           | Description     |
| ----------------------------- | --------------- |
| `SWAG_AWS_ACCESS_KEY_ID`      | Your aws access key id (public key). Used for image uploads. |
| `SWAG_AWS_SECRET_ACCESS_KEY`  | Your aws secret access key (private key). Used for image uploads. |

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


### Generating RSA Keys

5w4g-server uses JSON Web Tokens for authentication. To sign the tokens, you will need files
containing private and public keys. Private keys for the development and test runtime environments
have already been created for you and are included in the config folder. To run in production, you will
need to generate your own keys and keep them private. Do this with the following commands on a
unix-type system:

``` bash
openssl genrsa -out config/id.rsa 1024 
openssl rsa -in config/id.rsa -pubout > config/id.rsa.pub 
```

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


Response Formats
----------------

The server will always respond with json. In general, there are no wrappers around root object. A successful
response will be the object you requested or created, or an array of objects in the case of a get request.

### Successful Requests
Successful requests will always return a 200 code. The body of the response is a json object. Here's an example
of a successful request when creating an item:
```json
{
    "name": "Ice Cube Sticker",
    "imageUrl": "http://placehold.it/350x350",
    "price": 3,
    "description": "This sticker is really cool. Ice cold, actually.",
    "id": "k4FbclRdLYXVfELnndelad"
}
```

Here's an example of a successful sign in request. In this case, there is only one field:

```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbklkIjoiVGFXY25NOXIyN1h4OVpnQm5kZGJjMiIsImV4cCI6MTQyMTY5ODQyMSwiaWF0IjoxNDE5MTA2NDIxfQ.0_GtGwP3XGwcFIYnF2EcKNUbl3bRKRgYvWCCF89uxes"
}
```

For successful DELETE requests, the response will just be an empty object:

```json
{}
```

Finally, here's an example of getting all existing items (GET /items), which returns an array of objects instead
of a single object:

```json
[
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
```

### Validation Errors
If there are any server-side validation errors, the server will respond with a 422 code. Validation errors only occur
when there is a problem with a form. The body of the response will be a map of field name to the errors associated with
that field.  Here's an example:

```json
{
    "email": [
        "email is required.",
        "email must be correctly formatted."
    ],
    "password": [
        "password is required."
    ]
}
```

Note that there can be more than one error associated with a particular field.

### General Errors
All other errors will return with a different response code (see the "Response Codes" section below). The body of the response
will be a single key, "error", and it's value will be a string describing the error that occurred. Here are two examples:

```json
{
    "error": "dial tcp 127.0.0.1:6379: connection refused"
}
```

```json
{
    "error": "You need to be signed in to do that!"
}
```


REST Endpoints
--------------

#### POST `/admin_users/sign_in`
Purpose: Sign in an admin user (i.e. get a fresh, valid token)

URL Parameters: none

Body Parameters:
(fields with an asterisk are required)

| Field            | Description     |
| ---------------- | --------------- |
| email\*          | The admin user's email address. Must be properly formatted. |
| password\*       | The admin user's password. |


#### POST `/admin_users`
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

#### GET `/admin_users/:id`
**Requires Admin Authentication**

Purpose: Get an existing admin user

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the admin user you want to get |

Body Parameters: none

#### GET `/admin_users`
**Requires Admin Authentication**

Purpose: List all existing admin users

URL Parameters: none

Body Parameters: none

#### DELETE `/admin_users/:id`
**Requires Admin Authentication**

Purpose: Delete an existing admin user

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the admin user you want to delete |

Body Parameters: none

#### POST `/items`
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
| image\*          | An image file which will be used as the image for this item.  |

#### GET `/items/:id`

Purpose: Get a single existing item

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the item you want to get |

Body Parameters: none

#### GET `/items`

Purpose: List all existing items

URL Parameters: none

Body Parameters: none

#### DELETE `/items/:id`
**Requires Admin Authentication**

Purpose: Delete an existing item

URL Parameters:

| Field         | Description     |
| ------------- | --------------- |
| id\*          | The id of the item you want to delete |

Body Parameters: none

#### PUT `/items/:id`
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
| image         | An image file which will be used as the image for this item. |


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

**418: I am a Teapot**
This error code is returned iff the server has unintentionally turned into (or
perhaps gained control of) a teapot and you are attempting to brew coffee with it.
You may not attempt to brew coffee using the server under these circumstances. 

**422: Unprocessable Entity**  
The paramaters for the request were either incorrect or improperly formatted.
This is the response code for validation errors, and will typically include information
about which form fields were invalid. You should display the error(s) to the user.

**500: Internal Server Error**  
An error occured on the server-side. Will always return an errors field
in the JSON response which contains details about the error. If the runtime environment
is set to production, all internal server errors will simply contain the text: "Sorry
there was a problem." for security reasons.
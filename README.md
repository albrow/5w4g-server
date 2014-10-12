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

Getting Up and Running
----------------------

Install dependencies:

```bash
go get ./...
```

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

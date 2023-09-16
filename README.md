
# TritonHTTP

A web server that implements a subset of the HTTP/1.1 protocol, from scratch in Go. The server supports concurrent connections, persistent connections, HTTP pipelining, and virtual hosting. 

## Description

TritonHTTP follows the [general HTTP message format](https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages). It supports the following features:

- HTTP version supported: `HTTP/1.1`
- Request method supported: `GET`
- Response status supported:
  - `200 OK`
  - `400 Bad Request`
  - `404 Not Found`
- Request headers:
  - `Host` (required)
  - `Connection` (optional, `Connection: close` has special meaning influencing server logic)
  - Other headers are allowed, but won't have any effect on the server logic
- Response headers:
  - `Date` (required)
  - `Last-Modified` (required for a `200` response)
  - `Content-Type` (required for a `200` response)
  - `Content-Length` (required for a `200` response)
  - `Connection: close` (required in response for a `Connection: close` request, or for a `400` response)
  - Response headers should be written in sorted order for the ease of testing
  - Response headers should be returned in 'canonical form', meaning that the first letter and any letter following a hyphen should be upper-case. All other letters in the header string should be lower-case.

## Usage

The source code for tools needed to interact with TritonHTTP can be found in `cmd`. The following commands can be used to launch these tools:

1) `make fetch` - Allows you to construct custom responses and send them to the web server. Please refer to the README in `fetch`'s directory for more information.

2) `make gohttpd` - Starts up Go's inbuilt web-server.

3) `make tritonhttpd`  - Starts up TritonHTTP (localhost, port 8080).

Refer Makefile for more details.

## Tests

Some basic tests can be found in `cmd/tritonhttpd`.

## Source code

The source code for the web server can be found in `tritonhttp/`.

## Virtual Hosting

In some cases, it is desirable to host multiple web servers on a single physical machine. This allows all the hosted web servers to share the physical server’s resources such as memory and processing, and in particular, to share a single IP address. This project implements virtual hosting by allowing TritonHTTP to host multiple servers. Each of these servers has a unique host name and maps to a unique docroot directory (in `docroot_dirs/`) on the physical server. Every request sent to TritonHTTP includes the “Host” header, which is used to determine the web server that each request is destined for. 

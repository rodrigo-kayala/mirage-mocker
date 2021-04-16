# Mirage Mocker

**Mirage mocker** is a simple yet powerful API mock server. Mirage mocker can mock an HTTP request by:

* Returning same body received (request-response mock)
* Returning a fixed payload
* Running a customized response produced by a [Go Plugin](https://golang.org/pkg/plugin/)
* Transform the request (using a [Go Plugin](https://golang.org/pkg/plugin/)) and **proxy-pass** the HTTP requests to a real server.

## Installation

### Using go

Once you have [go installed](http://golang.org/doc/install.html#releases) and make sure to add $GOPATH/bin to your PATH.

#### Go version < 1.16

``` sh
GO111MODULE=on go get github.com/rodrigo-kayala/mirage-mocker@v1.1.0
 ```

#### Go version >= 1.16

``` sh
go install github.com/rodrigo-kayala/mirage-mocker@v1.1.0
 ```

### Using binary distribuition file

You can download the pre-build binary file compatible with your system [here](https://github.com/rodrigo-kayala/mirage-mocker/releases)

## Usage

Once you have the configuration file ready, simple run:

``` sh
mirage-mocker [config-file-path]
```

If you don't inform a configuration  file, mirage-mocker will look for a `mocker.yml` in current path

## Configuration

### Configuration example

This configuration is avaliable in [here](example.yml) and is the same configuration used on the tests cenarios. It also uses a [runnable](processor/testdata/runnable) and a [transform](processor/testdata/tranform) plugin

``` yaml
pretty-logs: true
services:
  - parser:
      pattern: /echo.*
      rewrite:
        - source: /echo(/.*)
          target: $1
      methods: [POST, GET, PUT, DELETE]
      type: pass
      log: true
      transform-lib: processor/testdata/transform/transform.so
      transform-symbol: AddHeader
      pass-base-uri: "https://postman-echo.com"
  - parser: 
      pattern: /.*$
      methods: [POST,PUT,DELETE]
      content-type: application/json
      type: mock
      log: true
      response:
        content-type: application/json
        status:
          POST: 201
          PUT: 200
          DELETE: 204
        body-type: request
  - parser: 
      pattern: /mock/fixed/value.*
      methods: [GET]
      type: mock
      log: true
      response:
        content-type: text/plain
        status:
          GET: 200
        body-type: fixed
        body: pong
  - parser:
      pattern: /mock/fixed/delay.*
      methods: [GET]
      type: mock
      log: true
      delay:
        min: 2s
        max: 3s
      response:
        content-type: text/plain
        status:
          GET: 200
        body-type: fixed
        body: pong
  - parser:
      pattern: /mock/fixed/file.*
      methods: [GET]
      type: mock
      log: true
      response:
        content-type: application/json
        status:
          GET: 200
        body-type: fixed
        body-file: processor/testdata/response1.json
  - parser:
      pattern: /mock/runnable.*
      methods: [GET]
      type: mock
      log: true
      response:
        status:
          GET: 200
        body-type: runnable
        response-lib: processor/testdata/runnable/runnable.so
        response-symbol: GetEnv
```

### Base attributes

* **pattern** *(required)*: regex pattern expression used to match requests URLs (without host)
* **methods** *(required)*: array of HTTP methods to match
* **type** *(required)*: *mock* or *pass* (proxy-pass)
* **content-type** *(optional)*: content-type to match (if the Content-Type header is not present or contains a different value, the request will not match)
* **log** *(optional): tells if request/response content should be logged. Defaults to **false**
* **delay** *(optional): adds a random delay to the request (could be useful to simulate real production cenarios)
  * **min**: min delay time that should added
  * **max**: max delay time that should added

Other attributes are specific for some type of request or response. See below.

### Mock - Base attributes

All mock type configurations should contains a response configuration

* **response** *(required for type **mock**)*
  * **status** *(required for matched methods)*
    * [*METHOD*]: [*HTTP RESPONSE STATUS CODE*]
    * ex. **GET**: 200
  * **body-type**: *fixed*, *request* or *runnable*

### Mock - fixed

Produces a fixed response for a given path/method. It can return a fixed string

```yaml
  - parser: 
      pattern: /mock/fixed/value.*
      methods: [GET]
      type: mock
      log: true
      response:
        content-type: text/plain
        status:
          GET: 200
        body-type: fixed
        body: pong

```

Or the content of a file:

```yaml
  - parser:
      pattern: /mock/fixed/file.*
      methods: [GET]
      type: mock
      log: true
      response:
        content-type: application/json
        status:
          GET: 200
        body-type: fixed
        body-file: processor/testdata/response1.json
```

#### Attributes

* **body**: response string

or

* **body-file**: file containing the response string

### Mock - request response

Response will always have same body as the request

```yaml
  - parser: 
      pattern: /.*$
      methods: [POST,PUT,DELETE]
      content-type: application/json
      type: mock
      log: true
      response:
        content-type: application/json
        status:
          POST: 201
          PUT: 200
          DELETE: 204
        body-type: request
```

### Mock - runnable

Run a customized *go plugin* which should perform the response

```yaml
  - parser:
      pattern: /mock/runnable.*
      methods: [GET]
      type: mock
      log: true
      response:
        status:
          GET: 200
        body-type: runnable
        response-lib: processor/testdata/runnable/runnable.so
        response-symbol: GetEnv
```

#### Attributes

* **response-lib**: Go plugin (*.so) file - *for instructions, se below*
* **response-symbol**: function to produce the response. It must have this signature:

```go
func (w  http.ResponseWriter, r *http.Request, status  int) error
```

[Here](processor/testdata/runnable/runnable.go) is a simple example of a *runnable* plugin

### Pass

Optionally runs a customized *go plugin* to transform request and then proxy-pass it to a real server.

```yaml
  - parser:
      pattern: /echo.*
      rewrite:
        - source: /echo(/.*)
          target: $1
      methods: [POST, GET, PUT, DELETE]
      type: pass
      log: true
      transform-lib: processor/testdata/transform/transform.so
      transform-symbol: AddHeader
      pass-base-uri: "https://postman-echo.com"
```

#### Attributes

* **rewrite**: a set of regex replaces to rewrite the URIs
  * **source**: regular expression
  * **target**: replace string
* **pass-base-uri**: base URI to proxy-pass requests
* **transform-lib**: Go plugin (*.so) file - *for instructions, se below*
* **transform-symbol**: function to transform the request. It must have this signature: `func (r *http.Request) error`

[Here](processor/testdata/transform/transform.go) is a simple example of a *transform* plugin

## Plugins

Mirage mocker plugins are standard Go plugins (see [https://golang.org/pkg/plugin](https://golang.org/pkg/plugin)).

### Runnable plugins

#### Function signature

```go
func (w http.ResponseWriter, r *http.Request, status int) error
```

#### Example

```go
func GetEnv(w http.ResponseWriter, r *http.Request, status int) error {
	var env string
	vname, ok := r.URL.Query()["vname"]
	if ok && len(vname) > 0 {
		env = os.Getenv(vname[0])
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf("%s", env)))

	return nil
}
```

### Transform plugins

#### Function signature

```go
func (r *http.Request) error {
```

#### Example

```go
func AddHeader(r *http.Request) error {
	vname, ok := r.URL.Query()["vname"]
	if ok && len(vname) > 0 {
		env := os.Getenv(vname[0])
		r.Header.Add(vname[0], env)
	}

	return nil
}
```

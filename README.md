# MIRAge Mocker
**MIRAge mocker** is a simple yet powerful API mock server. MIRAge mocker can mock any HTTP by:
* Returning same body received (request-response mock)
* Returning a fixed payload for a given Path
* Running a customized Go Plugin witch produces a response
Also, MIRAge mocker can transform (with customized Go Plugin) and **proxy-pass** your HTTP requests to a real server. 

## Installation

### Using go get
If you have go installed, simple run: 
``` bash
$ go get github.com/miraeducation/mirage-mocker
 ```
The binary file will be avaliable at `$GOPATH/bin`

### Using binary distribuition file
You can  simple download a binary distrubition file compatible with your system:
[https://github.com/miraeducation/mirage-mocker/releases](https://github.com/miraeducation/mirage-mocker/releases)

## Usage
Simple run `mirage-mocker`
``` bash
$ mirage-mocker [config-file-path]
```
If you don't inform a configuration  file, mirage-mocker will look for a `config.yml` in current path

## Configuration

### Configuration example
This configuration is avaliable in [here](https://github.com/miraeducation/mirage-mocker/blob/master/examples/config.yml) and is used for tests. Its also use a [runnable](https://github.com/miraeducation/mirage-mocker/tree/master/examples/plugins/runnable) plugin and a [transform](https://github.com/miraeducation/mirage-mocker/tree/master/examples/plugins/transform)
``` yaml
log-level: INFO
services:
  - parser: 
      pattern: /.*$
      methods: [POST,PUT,DELETE]
      content-type: application/json
      type: mock
      response:
        content-type: application/json
        status:
          POST: 201
          PUT: 200
          DELETE: 204
        body-type: request
  - parser: 
      pattern: /ping$
      methods: [GET]
      type: mock
      response:
        content-type: text/plain
        status:
          GET: 200
        body-type: fixed
        body: 'pong'
  - parser:
      pattern: /version$
      methods: [GET]
      type: mock
      response:
        status:
          GET: 200
        body-type: runnable
        response-lib: ../examples/plugins/runnable/runnable.so
        response-symbol: Version
  - parser:
      pattern: /test/pass$
      rewrite:
        - source: '/test(/.*)'
          target: '$1'
      methods: [GET]
      transform-lib: ../examples/plugins/transform/transform.so
      transform-symbol: AddHeader
      type: pass
      pass-base-uri: "http://localhost:55545"
```
### Global attributtes 
* **log-level**: global log level, using [logrus](https://github.com/sirupsen/logrus) log API.
### Parsers

#### Generic Attributes
* **pattern** *(required)*: regex pattern expression used to match requests URLs (without host)
*  **methods** *(required)*: Array of HTTP methods for match
* **content-type** *(optional)*: Content-type for match
* **type** *(required)*: *mock* or *pass* (proxy-pass)
* **response** *(required - only for type **mock**)*
	* **status** *(required for matched methods)*
		* [*METHOD*]: [*HTTP RESPONSE STATUS CODE*]
		* ex. **GET**: 200
	* **body-type**: *fixed*, *request* or *runnable*

Other attributes are specific for some type of request or response. See below.

#### Mock - fixed response
Fixed response for a given path/method
```yaml
pattern: /ping$
methods: [GET]
type: mock
response:
  content-type: text/plain
  status:
    GET: 200
  body-type: fixed
  body: 'pong'
```
##### Mock - Fixed  Type Attributes
* **body**: response string

#### Mock - request response
Response will have same body as request 
```yaml
 pattern: /.*$
 methods: [POST,PUT,DELETE]
 content-type: application/json
 type: mock
 response:
   content-type: application/json
   status:
     POST: 201
     PUT: 200
     DELETE: 204
   body-type: request
```

#### Mock - runnable
Run a customized  *go plugin* to respond 
```yaml
pattern: /version$
methods: [GET]
type: mock
response:
  status:
    GET: 200
  body-type: runnable
  response-lib: ../examples/plugins/runnable/runnable.so
  response-symbol: Version
```
##### Mock - Runnable  Type Attributes
* **response-lib**: Go plugin (*.so) file - *for instructions, se below*
* **response-symbol**: function witch will produce the response. Must have this signature: `func (w  http.ResponseWriter, r *http.Request, status  int) error` 

[Here](https://github.com/miraeducation/mirage-mocker/blob/master/examples/plugins/runnable/runnable.go)  is an example os *runnable* plugin

#### Pass
Run a customized  *go plugin* to transform request and proxy-pass to a real server 
```yaml
 pattern: /test/pass$
 rewrite:
   - source: '/test(/.*)'
     target: '$1'
 methods: [GET]
 transform-lib: ../examples/plugins/transform/transform.so
 transform-symbol: AddHeader
 type: pass
 pass-base-uri: "http://localhost:55545"
```
##### Mock - Runnable  Type Attributes
* **rewrite**: a set of regex replaces
	* **source**: regular expression
	* **target**: replace string
* **pass-base-uri**: base URI to proxy-pass requests
* **transform-lib**: Go plugin (*.so) file - *for instructions, se below*
* **transform-symbol**: function witch will transform the request. Must have this signature: `func (r *http.Request) error` 

[Here](https://github.com/miraeducation/mirage-mocker/blob/master/examples/plugins/transform/transform.go)  is an example os *transform* plugin

## Plugins
MIRAge mocker plugins are standard Go plugins (see [https://golang.org/pkg/plugin](https://golang.org/pkg/plugin)).
### Runnable plugins
#### Function signature
```go
func (w http.ResponseWriter, r *http.Request, status int) error
```
#### Example
```go
// Version writes current version response - runnable example
func Version(w http.ResponseWriter, r *http.Request, status int) error {
	version := "v1.0.0"
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(version))
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
// AddHeader adds an header to request
func AddHeader(r *http.Request) error {
	version := "v1.0.1"
	r.Header.Add("VERSION", version)
	return nil
}
```


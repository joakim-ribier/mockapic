# mockapic

[![Go Report Card](https://goreportcard.com/badge/github.com/joakim-ribier/mockapic)](https://goreportcard.com/report/github.com/joakim-ribier/mockapic)
![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/joakim-ribier/mockapic)

[![codecov](https://codecov.io/gh/joakim-ribier/mockapic/graph/badge.svg?token=AUAOC8992T)](https://codecov.io/gh/joakim-ribier/mockapic)
![example workflow](https://github.com/joakim-ribier/mockapic/actions/workflows/build-test-and-coverage.yml/badge.svg)
![example workflow](https://github.com/joakim-ribier/mockapic/actions/workflows/build-and-push-container.yml/badge.svg)

[![Docker Pulls](https://badgen.net/docker/pulls/joakimribier/mockapic?icon=docker&label=pulls)](https://hub.docker.com/r/joakimribier/mockapic/)
[![Docker Image Size](https://badgen.net/docker/size/joakimribier/mockapic?icon=docker&label=image%20size)](https://hub.docker.com/r/joakimribier/mockapic/)

`MOCKAPIC` is a Go HTTP server - The easiest way to test your web services securely and privately using a Docker container in Golang.

```bash
$ curl http://localhost:3333/


    __  ___              __                  _
   /  |/  /____   _____ / /__ ____ _ ____   (_)_____
  / /|_/ // __ \ / ___// //_// __ '// __ \ / // ___/
 / /  / // /_/ // /__ / ,<  / /_/ // /_/ // // /__   _  _  _
/_/  /_/ \____/ \___//_/|_| \__,_// .___//_/ \___/  (_)(_)(_)
                                 /_/
                    https://github.com/joakim-ribier/mockapic


 NAME                           | VALUE
                                |
 Requests max number authorized | unlimited
--------------------------------+--------------------------------------
 Remote addr total number       | 1
 Requests total number          | 10
                                |
 Last UUID                      | 00000000-0000-0000-0000-000000000000
 Last createdAt                 | yyyy-MM-dd hh:mm:ss


 List available APIs

 METHOD | ENDPOINT              | DESCRIPTION
        |                       |
 GET    | /                     | Get info
        +-----------------------+-------------------------------------
        | /static/content-types | Get allowed content types
        | /static/charsets      | Get allowed charsets
        | /static/status-codes  | Get allowed status codes
        +-----------------------+-------------------------------------
        | /v1/{uuid}            | Get a mocked request
        | /v1/list              | Get the list of all mocked requests
 POST   | /v1/add               | Create a new mocked request
```

[Usage](#usage) - [APIs](#apis) - [Test](#test) - [Docker](#docker) - [CI](#ci) - [Demo](#demo) - [License](#license)

## Usage

Deploy it as a service or use it directly in your development for integration tests.

| Option    | Env                     | Value                       | Default          | Description |
| ---       | ---                     | ---                         | ---              | ---
| --home    | MOCKAPIC_HOME             | /home/{user}/app/mockapic  | .                | Define the working directory<br>The folder must contain a `/requests` subfolder for the mocked requests
| --port    | MOCKAPIC_PORT             | 3333                        | 3333             | Define a specific port
| --req_max | MOCKAPIC_REQ_MAX_LIMIT    | 100                         | -1 (`unlimited`) | Define the max limit of the mocked requests
| --ssl     | MOCKAPIC_SSL              | true                        | false            | Enable SSL/Tls HTTP server (need to provide certificate)
| --cert    | MOCKAPIC_CERT             | /home/{user}/app/mockapic  | 3333             | Define the certificate directory to contain (`mockapic.cert` and `mockapic.key`)

1. Start the server

```bash
$ cd cmd/httpserver
$ ./httpserver --home /home/{user}/app/mockapic --port 3333
       ______        __  ___ ____   ______ __ ____  __      __   _____  ______ ____  _    __ ______ ____
      / ____/       /  |/  // __ \ / ____// //_/\ \/ /    _/_/  / ___/ / ____// __ \| |  / // ____// __ \
     / / __ ______ / /|_/ // / / // /    / ,<    \  /   _/_/    \__ \ / __/  / /_/ /| | / // __/  / /_/ /
    / /_/ //_____// /  / // /_/ // /___ / /| |   / /  _/_/     ___/ // /___ / _, _/ | |/ // /___ / _, _ /_  _  _
    \____/       /_/  /_/ \____/ \____//_/ |_|  /_/  /_/      /____//_____//_/ |_|  |___//_____//_/ |_| (_)(_)(_)
                                                                    https://github.com/joakim-ribier/mockapic
Server running on port 3333....
```

2. Create a new mocked request

```bash
$ curl --location 'http://localhost:3333/v1/new' \
--header 'Content-Type: application/json' \
--data '{
    "name": "Hello World - 200"
    "status": 200,
    "contentType": "text/plain",
    "charset": "UTF-8",
    "body": "Hello World",
    "headers": {
        "x-language": "golang",
        "x-host": "https://github.com/joakim-ribier/mockapic"
    }
}' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   272  100    48  100   224  30808   140k --:--:-- --:--:-- --:--:--  265k
{
  "uuid": "78c8e9a9-0ffe-48b9-abaa-4ca00d7eca6c"
}
```

3. Call it with the `uuid` value from the response

```bash
$ curl -v --location 'http://localhost:3333/v1/ca243375-8db6-4ff3-839b-16f9f90edc64'
> ...
< HTTP/1.1 200 OK
< Content-Type: text/plain; charset=UTF-8
< X-Host: https://github.com/joakim-ribier/mockapic
< X-Language: golang
Hello World
```

for more details see [APIs](#apis)

### SSL/Tls

Run the HTTP server in SSL/Tls (`https`) mode with certificate.

* `--ssl` (or `$MOCKAPIC_SSL`) must be `true`
* `--cert` (or `$MOCKAPIC_CERT`) must be contain a `mockapic.crt` and `mockapic.key` files

```bash
$ ./httpserver \
  --home /home/{user}/app/mockapic \
  --port 3333 \
  --ssl true \
  --cert {certificate-path-directory}
```

### How to use it in test

[joakim-ribier/go-utils - Full example](https://github.com/joakim-ribier/go-utils/blob/main/pkg/httpsutil/httpsutil_test.go)

```go
// uses a sensible default on windows (tcp/http) and linux/osx (socket)
pool, err := dockertest.NewPool("")
if err != nil {
  log.Fatalf("Could not construct pool: %s", err)
}

// uses pool to try to connect to Docker
err = pool.Client.Ping()
if err != nil {
  log.Fatalf("Could not connect to Docker: %s", err)
}

resource, err := pool.RunWithOptions(&dockertest.RunOptions{
  Repository:   "joakimribier/mockapic",
  Tag:          "latest",
  Env:          []string{"MOCKAPIC_PORT=3333"},
  ExposedPorts: []string{"3333"},
}, func(config *docker.HostConfig) {
  // set AutoRemove to true so that stopped container goes away by itself
  config.AutoRemove = true
  config.RestartPolicy = docker.RestartPolicy{Name: "no"}
})
if err != nil {
  log.Fatalf("Could not start resource: %s", err)
}

resource.Expire(30) // hard kill the container in 3 minutes (180 Seconds)
exposeHost = net.JoinHostPort("0.0.0.0", resource.GetPort("3333/tcp"))

if err := pool.Retry(func() error {
  req, err := NewHttpRequest(fmt.Sprintf("http://%s/", exposeHost), "")
  if err != nil {
    return err
  }
  _, err = req.Timeout("150ms").Call()
  return err
}); err != nil {
  log.Fatalf("Could not connect to mockapic server: %s", err)
}

code := m.Run()

// cannot defer this because os.Exit doesn't care for defer
if err := pool.Purge(resource); err != nil {
  log.Fatalf("Could not purge resource: %s", err)
}

os.Exit(code)
```

## APIs

List APIs available

| Method | Endpoint                              | Description |
| ---    | ---                                   | ---
| GET    | /                                     | Get info
| GET    | /static/content-types                 | Get allowed content types
| GET    | /static/charsets                      | Get allowed charsets
| GET    | /static/status-codes                  | Get allowed status codes
| GET    | [/v1/{uuid}](#get-mocked-request)     | Get a mocked request
| GET    | [/v1/list](#list-requests)            | Get the list of all mocked requests
| POST   | [/v1/add](#create-new-mocked-request) | Create a new mocked request

#### Create New Mocked Request

```bash
$ curl -X POST --location '~/v1/new' \
--header 'Content-Type: application/json' \
--data '{
    "status": 200,
    "contentType": "text/plain",
    "charset": "UTF-8",
    "body": "Hello World",
    "headers": {
        "x-language": "golang",
        "x-host": "https://github.com/joakim-ribier/mockapic"
    }
}' | jq
{
  "uuid": "{uuid}"
}
```

| Field       | Required | Value
| ---         | ---      | ---
| name        |          | Set a name to the request
| status      | [x]      | Code HTTP (`200`, `204`, `404`, ...)
| contentType | [x]      | Content Type (`application/json`, `text/plain`...)
| charset     | [x]      | Charset: `UTF-8`, `UTF-16` or `ISO-8859-1`
| body        | [x]      | Body returns by the request (`[]bytes(text, json)`)
| headers     |          | Header parameters (`x-key: value`)

#### Get Mocked Request

```bash
$ curl -v -X GET --location '~/v1/{uuid}?delay=100ms'
> ...
< HTTP/1.1 200 OK
< Content-Type: text/plain; charset=UTF-8
< X-Host: https://github.com/joakim-ribier/mockapic
< X-Language: golang
Hello World
```

| Field       | Required | Value
| ---         | ---      | ---
| {uuid}      | [x]      | Request ID returned by the POST API
| delay       |          | Parameter to the URL to delay the response. Maximum delay: `60s`

#### List requests

```bash
$ curl -X GET --location '~/v1/list' | jq
[
  {
    "Name": "200 - Ok",
    "UUID": "03e122b3-42d7-41bd-92ca-35b93ea38c4e",
    "CreatedAt": "1970-01-01 00:00:01",
    "Status": 200,
    "ContentType": "application/json"
  },
  {
    "Name": "409 - Conflict",
    "UUID": "051d6a25-33ea-49fe-86a7-07648489e750",
    "CreatedAt": "1970-01-01 00:00:01",
    "Status": 409,
    "ContentType": "application/json"
  },
  ...
```

## Test

```go
$ go test ./... -race -covermode=atomic -coverprofile=coverage.out
...
ok  	github.com/joakim-ribier/mockapic/internal	1.642s	coverage: 100.0% of statements
ok  	github.com/joakim-ribier/mockapic/internal/server	2.181s	coverage: 91.0% of statements
```

## Docker

### Pull and Run

```bash
# pull the last version
$ docker pull joakimribier/mockapic:latest

# run the docker image
$ docker run -it --restart unless-stopped -p 3333:3333 -e MOCKAPIC_PORT=3333 -e MOCKAPIC_REQ_MAX_LIMIT=100 -e MOCKAPIC_SSL=true \
  -v /home/{user}/app/mockapic:/usr/src/app/mockapic/data \
  -v /home/{user}/app/mockapic:/usr/src/app/mockapic/cert:ro joakimribier/mockapic
```

### Build

```bash
# `--platform linux/amd64` to build container for Github Action
$ docker build --platform linux/amd64 -t mockapic .

...
=> exporting to image
=> => exporting layers
=> => writing image sha256:0ffc9be6749b23b64cad0aa4b665ca26b9e53b649855a156b9325557620c57d1
=> => naming to docker.io/library/mockapic

$ docker run -it --rm -p 3333:3333 -e MOCKAPIC_PORT=3333 mockapic
...
Server running on port 3333....
```

### Push

```bash
$ docker tag mockapic:latest joakimribier/mockapic:latest
$ docker login -u "{username}" docker.io
$ docker push joakimribier/mockapic:latest
```

### Save and Load

```bash
$ docker save -o mockapic mockapic

$ scp ... # copy the image

$ docker load -i mockapic
...
dd63366eb47d: Loading layer [==================================================>]  77.58MB/77.58MB
3025ebe291d8: Loading layer [==================================================>]  7.267MB/7.267MB
Loaded image: mockapic:latest
```

## CI

### Github action workflows *.yml

|     | Name                                | Description
| --- | ---                                 | ---
|     | `build_test_and_coverage_reusable`  | Reusable workflow: build, execute test and push coverage (`on condition`)
| #1  | `build-test-and-coverage`           | Calls the reusable workflow on `main` branch
| #2  | `build-and-push-container`          | 1. Builds and pushes container on Docker Hub if the workflow `#1` is completed</br>2. Sends event to [joakim-ribier/go-utils](https://github.com/joakim-ribier/go-utils) to trigger action `trigger-from-event:build_and_test`
| #3  | `pr-test-only.yml`                  | Calls the reusable workflow on `pull request` (without coverage)

## Demo

Access to the demo version to try the service (`limited to 100 mocked requests`), feel free to use it [https://mockapic-dev](https://mockapic.dev).

_the server is not always operational so don't hesitate to try later_

## License

This software is licensed under the MIT license, see [License](https://github.com/joakim-ribier/go-utils/blob/main/LICENSE) for more information.
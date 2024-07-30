# gmocky-v2

[![Go Report Card](https://goreportcard.com/badge/github.com/joakim-ribier/gmocky-v2)](https://goreportcard.com/report/github.com/joakim-ribier/gmocky-v2)
![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/joakim-ribier/gmocky-v2)

[![codecov](https://codecov.io/gh/joakim-ribier/gmocky-v2/graph/badge.svg?token=AUAOC8992T)](https://codecov.io/gh/joakim-ribier/gmocky-v2)
![example workflow](https://github.com/joakim-ribier/gmocky-v2/actions/workflows/build-test-and-coverage.yml/badge.svg)
![example workflow](https://github.com/joakim-ribier/gmocky-v2/actions/workflows/build-and-push-container.yml/badge.svg)

[![Docker Pulls](https://badgen.net/docker/pulls/joakimribier/gmocky-v2?icon=docker&label=pulls)](https://hub.docker.com/r/joakimribier/gmocky-v2/)
[![Docker Image Size](https://badgen.net/docker/size/joakimribier/gmocky-v2?icon=docker&label=image%20size)](https://hub.docker.com/r/joakimribier/gmocky-v2/)

`GMOCKY-v2` is a Go HTTP server - The easiest way to test your web services securely and privately using a Docker container in Golang.

```bash
$ curl http://localhost:3333/

       ______        __  ___ ____   ______ __ ____  __      __   _____  ______ ____  _    __ ______ ____
      / ____/       /  |/  // __ \ / ____// //_/\ \/ /    _/_/  / ___/ / ____// __ \| |  / // ____// __ \
     / / __ ______ / /|_/ // / / // /    / ,<    \  /   _/_/    \__ \ / __/  / /_/ /| | / // __/  / /_/ /
    / /_/ //_____// /  / // /_/ // /___ / /| |   / /  _/_/     ___/ // /___ / _, _/ | |/ // /___ / _, _ /_  _  _
    \____/       /_/  /_/ \____/ \____//_/ |_|  /_/  /_/      /____//_____//_/ |_|  |___//_____//_/ |_| (_)(_)(_)
                                                                    https://github.com/joakim-ribier/gmocky-v2


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

[Usage](#usage) - [APIs](#apis) - [Test](#test) - [Docker](#docker) - [CI](#ci) - [License](#license)

## Usage

Deploy it as a service or use it directly in your development for integration tests.

| Option | Value                       | Default | Description |
| ---    | ---                         | ---     | ---
| --home | /home/{user}/data/gmocky-v2 | .       | Define the working directory or use `$GMOCKY_HOME` env
| --port | 3333                        | 3333    | Define a specific port or use `$GMOCKY_PORT` env

1. Start the server

```bash
$ cd cmd/httpserver
$ ./httpserver --home /home/{user}/data/gmocky-v2 --port 3333
#$ docker run -it --rm -p 3333:3333 -e GMOCKY_HOME=/home/{user}/data/gmocky-v2 -e GMOCKY_PORT=3333 gmocky-v2

       ______        __  ___ ____   ______ __ ____  __      __   _____  ______ ____  _    __ ______ ____
      / ____/       /  |/  // __ \ / ____// //_/\ \/ /    _/_/  / ___/ / ____// __ \| |  / // ____// __ \
     / / __ ______ / /|_/ // / / // /    / ,<    \  /   _/_/    \__ \ / __/  / /_/ /| | / // __/  / /_/ /
    / /_/ //_____// /  / // /_/ // /___ / /| |   / /  _/_/     ___/ // /___ / _, _/ | |/ // /___ / _, _ /_  _  _
    \____/       /_/  /_/ \____/ \____//_/ |_|  /_/  /_/      /____//_____//_/ |_|  |___//_____//_/ |_| (_)(_)(_)
                                                                    https://github.com/joakim-ribier/gmocky-v2
Server running on port 3333....
```

2. Create a new mocked request

```bash
$ curl --location 'http://localhost:3333/v1/new' \
--header 'Content-Type: application/json' \
--data '{
    "status": 200,
    "contentType": "text/plain",
    "charset": "UTF-8",
    "body": "Hello World",
    "headers": {
        "x-language": "golang",
        "x-host": "https://github.com/joakim-ribier/gmocky-v2"
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
< X-Host: https://github.com/joakim-ribier/gmocky-v2
< X-Language: golang
Hello World
```

for more details see [APIs](#apis)

### SSL/Tls

Run the HTTP server in SSL/Tls (`https`) mode with certificate.

* `--ssl` (or `$GMOCKY_SSL`) must be `true`
* `--cert` (or `$GMOCKY_CERT`) must be contain a `gmocky.crt` and `gmocky.key` files

```bash
$ ./httpserver \
  --home /home/{user}/data/gmocky-v2 \
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
  Repository:   "joakimribier/gmocky-v2",
  Tag:          "latest",
  Env:          []string{"GMOCKY_PORT=3333"},
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
  log.Fatalf("Could not connect to gmocky-v2 server: %s", err)
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
        "x-host": "https://github.com/joakim-ribier/gmocky-v2"
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
< X-Host: https://github.com/joakim-ribier/gmocky-v2
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
    "UUID": "03e122b3-42d7-41bd-92ca-35b93ea38c4e",
    "Status": 200,
    "ContentType": "application/json"
  },
  {
    "UUID": "051d6a25-33ea-49fe-86a7-07648489e750",
    "Status": 409,
    "ContentType": "text/plain"
  },
  ...
```

## Test

```go
$ go test ./... -race -covermode=atomic -coverprofile=coverage.out
...
ok  	github.com/joakim-ribier/gmocky-v2/internal	1.642s	coverage: 100.0% of statements
ok  	github.com/joakim-ribier/gmocky-v2/internal/server	2.181s	coverage: 91.0% of statements
```

## Docker

### Build

```bash
# `--platform linux/amd64` to build container for Github Action
$ docker build --platform linux/amd64 -t gmocky-v2 .

...
=> exporting to image
=> => exporting layers
=> => writing image sha256:0ffc9be6749b23b64cad0aa4b665ca26b9e53b649855a156b9325557620c57d1
=> => naming to docker.io/library/gmocky-v2

$ docker run -it --rm -p 3333:3333 -e GMOCKY_PORT=3333 gmocky-v2
...
Server running on port 3333....
```

### Push

```bash
$ docker tag gmocky-v2:latest joakimribier/gmocky-v2:latest
$ docker login -u "{username}" docker.io
$ docker push joakimribier/gmocky-v2:latest
```

### Save and Load

```bash
$ docker save -o gmocky-v2 gmocky-v2

$ scp ... # copy the image

$ docker load -i gmocky-v2
...
dd63366eb47d: Loading layer [==================================================>]  77.58MB/77.58MB
3025ebe291d8: Loading layer [==================================================>]  7.267MB/7.267MB
Loaded image: gmocky-v2:latest
```

## CI

### Github action workflows *.yml

|     | Name                                | Description
| --- | ---                                 | ---
|     | `build_test_and_coverage_reusable`  | Reusable workflow: build, execute test and push coverage (`on condition`)
| #1  | `build-test-and-coverage`           | Calls the reusable workflow on `main` branch
| #2  | `build-and-push-container`          | 1. Builds and pushes container on Docker Hub if the workflow `#1` is completed</br>2. Sends event to [joakim-ribier/go-utils](https://github.com/joakim-ribier/go-utils) to trigger action `trigger-from-event:build_and_test`
| #3  | `pr-test-only.yml`                  | Calls the reusable workflow on `pull request` (without coverage)

## License

This software is licensed under the MIT license, see [License](https://github.com/joakim-ribier/go-utils/blob/main/LICENSE) for more information.
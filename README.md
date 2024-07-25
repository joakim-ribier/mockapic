# gmocky-v2

[![Go Report Card](https://goreportcard.com/badge/github.com/joakim-ribier/gmocky-v2)](https://goreportcard.com/report/github.com/joakim-ribier/gmocky-v2)
![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/joakim-ribier/gmocky-v2)
[![codecov](https://codecov.io/gh/joakim-ribier/gmocky-v2/graph/badge.svg?token=AUAOC8992T)](https://codecov.io/gh/joakim-ribier/gmocky-v2)

`GMOCKY-v2` is a Go HTTP server - The easiest way to test your web services securely and privately using a Docker container in Golang.

[Usage](#usage) - [APIs](#apis) - [Test](#test) - [Docker](#docker) - [License](#license)

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

# or directly with the container

$ docker run -it --rm -p 3333:3333 -e GMOCKY_HOME=/home/{user}/data/gmocky-v2 -e GMOCKY_PORT=3333 gmocky-v2
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

## APIs

List APIs available

| Method | Endpoint                              | Description |
| ---    | ---                                   | ---
| GET    | /                                     | Get info
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

## License

This software is licensed under the MIT license, see [License](https://github.com/joakim-ribier/go-utils/blob/main/LICENSE) for more information.
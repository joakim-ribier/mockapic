# mockapic

[![Go Report Card](https://goreportcard.com/badge/github.com/joakim-ribier/mockapic)](https://goreportcard.com/report/github.com/joakim-ribier/mockapic)
![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/joakim-ribier/mockapic)

[![codecov](https://codecov.io/gh/joakim-ribier/mockapic/graph/badge.svg?token=AUAOC8992T)](https://codecov.io/gh/joakim-ribier/mockapic)
![example workflow](https://github.com/joakim-ribier/mockapic/actions/workflows/build-test-and-coverage.yml/badge.svg)
![example workflow](https://github.com/joakim-ribier/mockapic/actions/workflows/build-and-push-container.yml/badge.svg)

[![Docker Pulls](https://badgen.net/docker/pulls/joakimribier/mockapic?icon=docker&label=pulls)](https://hub.docker.com/r/joakimribier/mockapic/)
[![Docker Image Size](https://badgen.net/docker/size/joakimribier/mockapic?icon=docker&label=image%20size)](https://hub.docker.com/r/joakimribier/mockapic/)

`Mockapic` [mokapik] is a Go HTTP server - The easiest way to test your web services securely and privately using a Docker container.

It's always complicated to test easly your application when it uses external API services, for many reasons:

* the service may not be free
* it is limited on the calls
* API response is not safe, it may change
* I don't want to expose my authentication data (API_KEY) in my tests
* it's not possible to test my implementation on several behaviors
* ...

It's for all these reasons that I created `Mockapic` (based on the awesome [mocky.io](https://designer.mocky.io/) idea). The main goal is to be able to better test your code on several behaviors when calling external API services.

[Usage](#usage) - [How it works](#how-it-works) - [APIs](#apis) - [Test](#test) - [Docker](#docker) - [CI](#ci) - [Demo](#demo) - [Thanks](#thanks) - [License](#license)

## Usage

Use it as a service or directly with Docker.

### As a service

```bash
# install the latest version
$ go install -v github.com/joakim-ribier/mockapic/cmd/httpserver@latest

# use the service
$ httpserver --port 3333 --home /home/{user}/app/mockapic
```

### As a container

```bash
# pull the last version
$ docker pull joakimribier/mockapic:latest

# run the image
$ docker run -it --rm -p 3333:3333 -e MOCKAPIC_PORT=3333 joakimribier/mockapic
```

## How it works

So, let's say I want to test my service which converts an EUR amount to USD. To convert the amount, my service needs to call an external API which returns the latest exchange rates data on the world.

To test my service, I want to mock the response of the external API to simulate several behaviors depends on the API result.

1. I have to start `Mockapic` server (see the all parameters)
2. I need to create several mocked requests based on the external API for different use case
3. I have to configure my test to call the mocked url test instead of the external API

Parameters of the service

| Option    | Env                     | Value                       | Default          | Description |
| ---       | ---                     | ---                         | ---              | ---
| --home    | MOCKAPIC_HOME           | /usr/app/mockapic           | .                | Define the working directory
| --port    | MOCKAPIC_PORT           | 3333                        | 3333             | Define a specific port
| --req_max | MOCKAPIC_REQ_MAX_LIMIT  | 100                         | -1 (`unlimited`) | Define the max limit of the mocked requests
| --ssl     | MOCKAPIC_SSL            | true                        | false            | Enable SSL/Tls HTTP server (need to provide certificate)
| --cert    | MOCKAPIC_CERT           | /usr/app/mockapic           | .                | Define the certificate directory which should contain (`mockapic.cert` and `mockapic.key`)

1. Start `Mockapic`

```bash
$ docker run -it -p 3333:3333 -e MOCKAPIC_PORT=3333 mockapic &
    __  ___              __                  _
   /  |/  /____   _____ / /__ ____ _ ____   (_)_____
  / /|_/ // __ \ / ___// //_// __ '// __ \ / // ___/
 / /  / // /_/ // /__ / ,<  / /_/ // /_/ // // /__   _  _  _
/_/  /_/ \____/ \___//_/|_| \__,_// .___//_/ \___/  (_)(_)(_)
                                 /_/
                    https://github.com/joakim-ribier/mockapic

Server running on port http[:3333]....
```

2. Create new mocked requests

The first one simulate a real response of the external API service which returns the exchange rates:

```bash
$ curl -X POST 'http://localhost:3333/v1/new?status=200&contentType=application%2Fjson&charset=UTF-8&domain=github.com%2Fjoakim-ribier&project=mockapic' \
--data '{
  "timestamp":1725967696,
  "rates":{
    "EUR":1,
    "GBP":0.842772,
    "KZT":527.025041,
    "USD":1.103546
  }
}' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   408  100    48  100   360  13892   101k --:--:-- --:--:-- --:--:--  132k
{
  "id": "c1403100-3aa0-484f-8e0f-f2c1db80f371",
  "_links": {
    "raw": "http://localhost:3333/v1/raw/c1403100-3aa0-484f-8e0f-f2c1db80f371",
    "self": "http://localhost:3333/v1/c1403100-3aa0-484f-8e0f-f2c1db80f371"
  }
}
```

And the second simulate an internal server error from the external API service:

```bash
$ curl -X POST 'http://localhost:3333/v1/new?status=500&contentType=application%2Fjson&charset=UTF-8&domain=github.com%2Fjoakim-ribier&project=mockapic' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   408  100    48  100   360  13892   101k --:--:-- --:--:-- --:--:--  132k
{
  "id": "79090265-a1af-47ec-a177-88668582ce28",
  "_links": {
    "raw": "http://localhost:3333/v1/raw/79090265-a1af-47ec-a177-88668582ce28",
    "self": "http://localhost:3333/v1/79090265-a1af-47ec-a177-88668582ce28"
  }
}
```

3. Use the `id` from the response to call the mocked URL
```bash
$ curl -GET 'http://localhost:3333/v1/c1403100-3aa0-484f-8e0f-f2c1db80f371' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    89  100    89    0     0   1226      0 --:--:-- --:--:-- --:--:--  1236
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=UTF-8
< Domain: github.com/joakim-ribier
< Project: mockapic
{
  "timestamp": 1725967696,
  "rates": {
    "EUR": 1,
    "GBP": 0.842772,
    "KZT": 527.025041,
    "USD": 1.103546
  }
}
```

or

```bash
$ curl -v -GET 'http://localhost:3333/v1/79090265-a1af-47ec-a177-88668582ce28'
* Request completely sent off
< HTTP/1.1 500 Internal Server Error
< Content-Type: application/json; charset=UTF-8
< Domain: github.com/joakim-ribier
< Project: mockapic
```

For more examples, please find for a specific language the use case of how to use `Mockapic` in your implementation:

* [`Golang`](https://github.com/joakim-ribier/mockapic-example-go)

### Predefined requests

If you don't want to create a new mocked requests every time, you can also version a mocked requests file and load it directly by the server on startup.

The file must be in `{MOCKAPIC_HOME}/mockapic.json` or directly mount to the right destination `-v /home/{user}/app/mockapic/mockapic.json:/usr/app/mockapic/mockapic.json` using as a Docker container.

See an example of [`mockapic.json`](/cmd/httpserver/mockapic.json) file

### SSL/Tls

Run the HTTP server in SSL/Tls (`https`) mode with certificate.

* `--ssl` parameter or `$MOCKAPIC_SSL` environment must be `true`
* `--cert` parameter or `$MOCKAPIC_CERT` environment must contain `mockapic.crt` and `mockapic.key` files

```bash
$ ./httpserver \
  --home /home/{user}/app/mockapic \
  --port 3333 \
  --ssl true \
  --cert /home/{user}/app/mockapic # by default --home directory
```

## APIs

List APIs available

| Method | Endpoint                              | Description |
| ---    | ---                                   | ---
| GET    | /                                     | Get info
| GET    | /static/content-types                 | Get allowed content types
| GET    | /static/charsets                      | Get allowed charsets
| GET    | /static/status-codes                  | Get allowed status codes
| GET    | [/v1/{id}](#get-mocked-request)       | Get a mocked request
| GET    | [/v1/raw/{id}](#raw-mocked-request)   | Get a raw mocked request
| GET    | [/v1/list](#list-requests)            | Get the list of all mocked requests
| POST   | [/v1/new](#create-new-mocked-request) | Create a new mocked request

#### Create New Mocked Request

```bash
$ curl -X POST '~/v1/new?status={status}&contentType={contentType}&charset={charset}&{header1}={header1}&{header2}={header2}' \
--data 'Hello World' | jq
{
  "id": "{id}",
  "_links": {
    "raw": "{host}/v1/raw/{id}",
    "self": "{host}/v1/{id}"
  }
}
```

| Field       | Required | Value
| ---         | ---      | ---
| status      | [x]      | Code HTTP (`200`, `204`, `404`, ...)
| contentType | [x]      | Content Type (`application/json`, `text/plain`...)
| charset     | [x]      | Charset: `UTF-8`, `UTF-16` or `ISO-8859-1`
| body        |          | Body returns by the request (`[]bytes(text, json)`)
| headers     |          | Header parameters (`x-key: value`)

#### Get Mocked Request

```bash
$ curl -X GET '~/v1/{id}?delay=100ms'
```

| Field       | Required | Value
| ---         | ---      | ---
| {id}        | [x]      | Request identifier returned by the POST API
| delay       |          | Parameter to the URL to delay the response - Maximum delay: `60s`

#### Raw Mocked Request

```bash
$ curl -X GET '~/v1/raw/{id}'

{
  "id": "{id}",
  "createdAt": "1970-01-01 00:00:01",
  "status": {status},
  "contentType": "{contentType}",
  "charset": "UTF-8",
  "headers": {
    "domain": "github.com/joakim-ribier",
    "project": "mockapic"
  },
  "body": "{raw}",
  "body64": "{base64}"
}
```

#### List requests

```bash
$ curl -X GET '~/v1/list' | jq
[
  {
    "id": "{id}",
    "createdAt": "1970-01-01 00:00:01",
    "status": {status},
    "contentType": "{contentType}",
    "charset": "UTF-8",
    "headers": {
      "domain": "github.com/joakim-ribier",
      "project": "mockapic"
    },
    "_links": {
      "raw": "{host}/v1/raw/{id}",
      "self": "{host}/v1/{id}"
    }
  },
  {
    "id": "{id}",
    "createdAt": "1970-01-01 00:00:01",
    "status": {status},
    "contentType": "{contentType}",
    "charset": "UTF-8",
    "headers": {
      "domain": "github.com/joakim-ribier",
      "project": "mockapic"
    },
    "_links": {
      "raw": "{host}/v1/raw/{id}",
      "self": "{host}/v1/{id}"
    }
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

# run the docker image with parameters
$ docker run -it --rm \
    -p 3333:3333 -e MOCKAPIC_PORT=3333 \
    -e MOCKAPIC_REQ_MAX_LIMIT=100 -e MOCKAPIC_SSL=true \
    -v /home/{user}/app/mockapic:/usr/app/mockapic \
    -v /home/{user}/app/mockapic/mockapic.json:/usr/app/mockapic/mockapic.json \
    joakimribier/mockapic
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

## CI

### Github action workflows *.yml

| Name | Description
| ---  | ---
| [Build test and coverage (reusable)](.github/workflows/build_test_and_coverage_reusable.yml) | It builds, executes tests and pushes coverage (`{input.coverage}`).
| [Build and push docker image to Docker Hub (reusable)](.github/workflows/build-and-push-container_reusable.yml) | It builds and pushes a container image on specific version (`latest` or `{input.version}`).
| [Build and test pull request](.github/workflows/pr-test-only.yml) | It calls `Build test and coverage (reusable)` workflow on all pull request without coverage.
| [Build, test and push coverage](.github/workflows/build-test-and-coverage.yml) | It calls `Build test and coverage (reusable)` workflow on the `main` branch with coverage.
| [Build and push docker image to Docker Hub](.github/workflows/build-and-push-container.yml) | It is launched after a `Build, test and push coverage` completed workflow on the `main` branch.
| [Publish binaries and tag Docker image](.github/workflows/release.yml)  | It is launched after a release creation, it builds binaries and calls `Build and push docker image to Docker Hub (reusable)` on a specific release tag version.
| [Publish new container image tag](.github/workflows/tag-container.yml)  | A manual workflow that calls `Build and push docker image to Docker Hub (reusable)` on a specific version.

## Demo

Access to the demo version to try the service (`limited to 100 mocked requests`), feel free to use it [https://mockapic-dev](https://gmocky.dev).

_the server is not always operational so don't hesitate to try later_

## Thanks

* [Bruno Adele](https://x.com/jesuislibre) and [Paul Leclercq](https://github.com/polomarcus) to help me to improve `Mockapic`
* [Julien Lafont](https://x.com/julien_lafont) for the awesome [mocky.io](https://designer.mocky.io/)

## License

This software is licensed under the MIT license, see [License](https://github.com/joakim-ribier/go-utils/blob/main/LICENSE) for more information.
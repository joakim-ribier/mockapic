FROM golang:1.22

RUN mkdir -p /usr/src/app/mockapic
RUN mkdir -p /usr/src/app/mockapic/data
RUN mkdir -p /usr/src/app/mockapic/data/requests
RUN mkdir -p /usr/src/app/mockapic/cert

WORKDIR /usr/src/app/mockapic

# build application
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...

# set ENV variables
ENV MOCKAPIC_PORT 3333
ENV MOCKAPIC_HOME /usr/src/app/mockapic/data
ENV MOCKAPIC_CERT /usr/src/app/mockapic/cert
# -1 / unlimited
ENV MOCKAPIC_REQ_MAX_LIMIT -1

 # if true the *.crt and *.key files must be provided
ENV MOCKAPIC_SSL false
# copy certificates (*.crt, *.key) files
COPY cert/* /usr/src/app/mockapic/cert/

CMD ["./httpserver"]
FROM golang:1.22

# source
RUN mkdir -p /usr/src/app/mockapic

# data
RUN mkdir -p /usr/app/mockapic
RUN mkdir -p /usr/app/mockapic/requests

WORKDIR /usr/src/app/mockapic

# build application
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...

# set ENV variables
ENV MOCKAPIC_PORT=3333
ENV MOCKAPIC_HOME=/usr/app/mockapic
ENV MOCKAPIC_CERT=/usr/app/mockapic
# -1 / unlimited
ENV MOCKAPIC_REQ_MAX_LIMIT=-1

 # if true the *.crt and *.key files must be provided
ENV MOCKAPIC_SSL=false

# copy example requests file in the default home directory
COPY cmd/httpserver/mockapic.json /usr/app/mockapic/

CMD ["./httpserver"]
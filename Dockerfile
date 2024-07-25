FROM golang:1.22

RUN mkdir -p /usr/src/app/gmocky-2
RUN mkdir -p /usr/src/app/gmocky-2/data

WORKDIR /usr/src/app/gmocky-2

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...

ENV GMOCKY_PORT 3333
ENV GMOCKY_HOME /usr/src/app/gmocky-2/data

CMD ["./httpserver"]
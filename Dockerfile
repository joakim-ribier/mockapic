FROM golang:1.22

RUN mkdir -p /usr/src/app/gmocky-2
RUN mkdir -p /usr/src/app/gmocky-2/data

WORKDIR /usr/src/app/gmocky-2

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...
RUN cp -a httpserver /usr/local/bin/

ENV GMOCKY_PORT 3333
ENV GMOCKY_HOME /usr/src/app/gmocky-2/data

CMD ["httpserver"]
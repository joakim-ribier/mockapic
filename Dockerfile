FROM golang:1.22

RUN mkdir -p /usr/src/app/gmocky-2
RUN mkdir -p /usr/src/app/gmocky-2/data
RUN mkdir -p /usr/src/app/gmocky-2/cert

WORKDIR /usr/src/app/gmocky-2

# build application
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...

# set ENV variables
ENV GMOCKY_PORT 3333
ENV GMOCKY_HOME /usr/src/app/gmocky-2/data
ENV GMOCKY_CERT /usr/src/app/gmocky-2/cert

 # if true the *.crt and *.key files must be provided
ENV GMOCKY_SSL false
# copy certificates (*.crt, *.key) files
COPY cert/* /usr/src/app/gmocky-2/cert/

CMD ["./httpserver"]
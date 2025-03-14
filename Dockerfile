### Step 1: Build stage
FROM golang:1.23.7-alpine AS builder

# source
RUN mkdir -p /usr/src/app/mockapic

WORKDIR /usr/src/app/mockapic

# build application
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...

###
## Step 2: Runtime stage
FROM scratch

# set ENV variables
ENV MOCKAPIC_PORT=3333
ENV MOCKAPIC_HOME=/usr/app/mockapic
ENV MOCKAPIC_CERT=/usr/app/mockapic
# -1 / unlimited
ENV MOCKAPIC_REQ_MAX_LIMIT=-1

 # if true the *.crt and *.key files must be provided
ENV MOCKAPIC_SSL=false

COPY --from=builder /usr/src/app/mockapic/httpserver /mockapic
COPY --from=builder /usr/src/app/mockapic/mockapic.json /usr/app/mockapic/

CMD ["./mockapic"]

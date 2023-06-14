FROM golang:1.20

# build directory will not be saved in the volume (which is mounted on /app)
RUN mkdir /build

WORKDIR /app
ADD . .
RUN go mod download
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -build="go build -o /build/server cmd/server/main.go" -command="/build/server"
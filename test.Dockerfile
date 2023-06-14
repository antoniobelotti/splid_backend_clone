FROM golang:1.20

# build directory will not be saved in the volume (which is mounted on /app)
RUN mkdir /build

WORKDIR /app
ADD . .
RUN go mod download

CMD go test -v ./...
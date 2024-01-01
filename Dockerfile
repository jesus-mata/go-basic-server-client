
# Build the application from source
FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /server

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM scratch AS build-release-stage

# import curl from curl scratch repository image
COPY --from=ghcr.io/tarampampam/curl:8.0.1 /bin/curl /bin/curl

WORKDIR /

COPY --from=build-stage /server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
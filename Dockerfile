
###### Build the application from source
FROM golang:1.21-bookworm AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /server

# Create a non-root user to run the application inside the container
# This is a best practice for security purposes (see https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user)
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 60000 \
    runner


####### Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...


###### Deploy the application binary into a lean image
FROM scratch AS build-release-stage

# Import the user and group files from the builder as well as the CA certificates and timezone files
COPY --from=build-stage /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-stage /etc/passwd /etc/passwd
COPY --from=build-stage /etc/group /etc/group

WORKDIR /

COPY --from=build-stage /server /server

EXPOSE 8080

# Run as the new non-root by default
USER runner:runner

ENTRYPOINT ["/server"]
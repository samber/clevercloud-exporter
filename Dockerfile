
FROM golang:1.20 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /clevercloud-exporter *.go

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /clevercloud-exporter /clevercloud-exporter

EXPOSE 9217

USER nonroot:nonroot

ENTRYPOINT ["/clevercloud-exporter"]

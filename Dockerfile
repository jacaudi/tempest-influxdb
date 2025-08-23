FROM golang:1.25-alpine AS build

WORKDIR /src/
COPY . /src/
RUN apk --no-cache add ca-certificates; \
    CGO_ENABLED=0 go build -o /bin/tempest-influx ./cmd/tempest-influx

FROM scratch
COPY --from=build /bin/tempest-influx /bin/tempest-influx
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 50222/udp

VOLUME "/config"

ENTRYPOINT ["/bin/tempest-influx"]

FROM golang:1.25-alpine AS build

ARG HTTP_PROXY HTTPS_PROXY NO_PROXY

COPY proxy-ca.crt* /usr/local/share/ca-certificates/
RUN if [ -f /usr/local/share/ca-certificates/proxy-ca.crt ]; then \
      cat /usr/local/share/ca-certificates/proxy-ca.crt >> /etc/ssl/certs/ca-certificates.crt; \
    fi

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /echo ./cmd/echo
RUN mkdir /data-dir

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /echo /echo
COPY --from=build --chown=nonroot:nonroot /data-dir /data
USER nonroot:nonroot

ENTRYPOINT ["/echo"]

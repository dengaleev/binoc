FROM golang:1.23-alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /echo ./cmd/echo

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /echo /echo
USER nonroot:nonroot

ENTRYPOINT ["/echo"]

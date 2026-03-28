FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /echo ./cmd/echo

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /echo /echo
USER nonroot:nonroot

ENTRYPOINT ["/echo"]

FROM golang:1.23.4-alpine as go-builder
RUN apk add --update make git
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
COPY cmd /src/cmd
WORKDIR /src
RUN CGO_ENABLED=0 go build ./...

FROM gcr.io/distroless/static
LABEL org.opencontainers.image.source https://github.com/prymitive/kthxbye
COPY --from=go-builder /src/kthxbye /kthxbye
EXPOSE 8080
ENTRYPOINT ["/kthxbye"]

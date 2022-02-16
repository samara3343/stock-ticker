FROM golang:1.17-alpine as builder

ADD . /app
WORKDIR /app
RUN go build -mod=vendor -o=/app/bin/stock-ticker .

FROM alpine:3

COPY --from=builder /app/bin/stock-ticker /app/bin/stock-ticker

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S app && adduser -S app -G app
RUN chown -R app:app /app
USER app
WORKDIR /app

ENTRYPOINT ["/app/bin/stock-ticker"]
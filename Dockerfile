FROM golang:1.21.6-alpine3.19 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN GOOS=linux go build -o bin/api main.go

# ----------------------------

FROM alpine

WORKDIR /app
COPY --from=builder /app/bin/api .
COPY --from=builder /app/.env .

RUN ls -lah

ENTRYPOINT [ "./api" ]
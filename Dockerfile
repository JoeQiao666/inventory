FROM golang:1.16-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build ./cmd/server.go

FROM alpine
COPY --from=build /app/server .
EXPOSE 8080
CMD [ "./server" ]
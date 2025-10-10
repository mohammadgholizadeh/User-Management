FROM golang:1.24.2-alpine AS build     

WORKDIR /user
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user ./cmd/server



FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /user
COPY --from=build /user/user /usr/local/bin/user
COPY --from=build /user/migrations /usr/local/bin/migrations 
COPY --from=build /user/configs /user/configs
COPY --from=build /user/docs   /user/docs
EXPOSE 8080
ENTRYPOINT ["user"]
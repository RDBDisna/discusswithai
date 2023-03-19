FROM golang:latest as builder

ARG GIT_COMMIT
ENV GIT_COMMIT=$GIT_COMMIT

WORKDIR /discusswithai

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .


RUN go get github.com/swaggo/swag/gen@v1.8.7
RUN go get github.com/swaggo/swag/cmd/swag@v1.8.7
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.7

RUN swag init

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$GIT_COMMIT" -o /bin/discusswithai .

FROM alpine:latest

RUN addgroup -S discusswithai && adduser -S discusswithai -G discusswithai

USER discusswithai
WORKDIR /home/discusswithai

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
COPY --from=builder /bin/discusswithai ./

ENV ZONEINFO=/zoneinfo.zip

EXPOSE 8000

ENTRYPOINT ["./discusswithai", "--dotenv=false"]

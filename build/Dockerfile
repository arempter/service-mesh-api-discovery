FROM golang:latest as build

ADD . /app

WORKDIR /app

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/receiver cmd/receiver.go

FROM alpine

RUN mkdir /app

COPY --from=build /app/receiver /app/receiver


RUN chmod +x /app/receiver

EXPOSE 65000


ENTRYPOINT [ "/app/receiver" ]
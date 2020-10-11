FROM golang:latest
WORKDIR /api
ADD . .
WORKDIR /api/cmd/app
RUN go build -o chat
ENTRYPOINT CONFIG_FILE=/api/config/docker-config.yml ./chat
EXPOSE 8000

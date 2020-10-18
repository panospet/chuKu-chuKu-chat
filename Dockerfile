FROM golang:latest
WORKDIR /api
RUN apt-get update
RUN apt-get install -y wamerican
ADD . .
WORKDIR /api/cmd/app
RUN go build -o chat
ENTRYPOINT CONFIG_FILE=/api/config/docker-config.yml ./chat
EXPOSE 8000

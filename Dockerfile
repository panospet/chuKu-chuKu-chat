FROM golang:1.21.0

WORKDIR /api
RUN apt update
RUN apt install -y wamerican
ADD . .
WORKDIR /api/cmd/app
RUN go build -o chat
ENTRYPOINT ./chat
EXPOSE 8000

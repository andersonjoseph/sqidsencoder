FROM golang:1.22

WORKDIR /usr/src/app

ENV PATH "$PATH:/usr/src/bin"
ENV SHELL "bash"

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

# install dlv (https://github.com/go-delve/delve)
RUN go install github.com/go-delve/delve/cmd/dlv@v1.22.1

FROM golang:1.15 AS base

ENV GO111MODULE on

FROM base AS testing

RUN GOPATH=/go \
    go get github.com/onsi/ginkgo/ginkgo@latest \
           github.com/onsi/gomega@latest \
           github.com/jackc/pgx@latest

FROM testing AS vscode

RUN GOPATH=/go \
    go get github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest \
           github.com/ramya-rao-a/go-outline@latest \
           github.com/go-delve/delve/cmd/dlv@latest \
           golang.org/x/lint/golint@latest \
           golang.org/x/tools/gopls@latest \
           github.com/gin-gonic/gin@latest

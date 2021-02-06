FROM golang:1.15 AS base

ENV GO111MODULE on

RUN GOPATH=/go \
    go get github.com/gin-gonic/gin@latest \
           github.com/jackc/pgx@latest

RUN GOPATH=/go \
    go get github.com/githubnemo/CompileDaemon@latest

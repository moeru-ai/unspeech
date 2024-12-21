list:
  @just --list

build:
  CGO_ENABLED=0 go build -o ./result/unspeech ./cmd/unspeech
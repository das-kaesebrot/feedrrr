FROM docker.io/library/golang:alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build

ARG VERSION="dev-docker"
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# https://jerrynsh.com/3-easy-ways-to-add-version-flag-in-go/
RUN go build -v -ldflags "-X 'main.Version=${VERSION}'" -o /usr/local/bin/app ./cmd/feedrrr/main.go

FROM scratch

ARG RUN_UID="10021"
ARG RUN_USER="feedrrr"

COPY --from=build /usr/local/bin/app feedrrr
COPY contrib/passwd /etc/passwd
COPY contrib/config.empty.yml /etc/feedrrr/config.yml
USER ${RUN_USER}

CMD ["./feedrrr"]

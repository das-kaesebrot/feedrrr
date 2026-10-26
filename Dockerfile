FROM docker.io/library/golang:alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS build

ARG VERSION="v0.0.1-docker"
ARG GIT_HASH="0000000000000000000000000000000000000000"
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# https://jerrynsh.com/3-easy-ways-to-add-version-flag-in-go/
RUN go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.GitHash=${GIT_HASH}'" -o /usr/local/bin/app ./cmd/feedrrr/main.go

FROM docker.io/library/alpine@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

ARG APP_WORKDIR="/var/opt/feedrrr"
ARG RUN_UID="10021"
ARG RUN_USER="feedrrr"

RUN apk add --no-cache tzdata
RUN mkdir -pv "${APP_WORKDIR}"
RUN addgroup -g ${RUN_UID} ${RUN_USER} && \
    adduser -h ${APP_WORKDIR} -u ${RUN_UID} -G ${RUN_USER} -s /bin/false -D ${RUN_USER} && \
    chown -R ${RUN_USER}:${RUN_USER} "${APP_WORKDIR}"
WORKDIR ${APP_WORKDIR}

COPY --from=build /usr/local/bin/app /usr/local/bin/feedrrr
COPY contrib/config.empty.yml /etc/feedrrr/config.yml
USER ${RUN_USER}

CMD ["feedrrr"]

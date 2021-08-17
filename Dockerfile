FROM golang:1.16-alpine AS build

RUN apk add --no-cache bash build-base git tree curl
WORKDIR /src

# dowload debugger into Docker cacheable layer
ENV GOBIN=/bin
ENV ROOT_DIR=/src

# download dependencies into Docker cacheable layer
COPY go.mod go.sum ./
RUN go mod download

# generate & build
ARG VERSION
ARG COMMIT
COPY . .
RUN go run mage.go deps build

FROM alpine
ARG VERSION
ARG COMMIT
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.source=https://github.com/aserto-dev/sver
LABEL org.opencontainers.image.title="sver version calculator"
LABEL org.opencontainers.image.revision=$COMMIT
LABEL org.opencontainers.image.url=https://aserto.com

RUN apk add --no-cache bash git openssh
WORKDIR /app
COPY --from=build /src/dist/build_linux_amd64/sver /app/

ENTRYPOINT ["./sver"]

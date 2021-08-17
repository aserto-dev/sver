FROM golang:1.16-alpine AS build
ARG SSH_PRIVATE_KEY
RUN apk add --no-cache bash build-base git tree curl protobuf openssh
WORKDIR /src

# make sure git ssh is properly setup so we can access private repos
RUN mkdir -p $HOME/.ssh && umask 0077 && echo -e "${SSH_PRIVATE_KEY}" > $HOME/.ssh/id_rsa \
	&& git config --global url."git@github.com:".insteadOf https://github.com/ \
	&& ssh-keyscan github.com >> $HOME/.ssh/known_hosts

# dowload debugger into Docker cacheable layer
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOBIN=/bin
ENV GOPRIVATE=github.com/aserto-dev
ENV ROOT_DIR=/src

# download dependencies into Docker cacheable layer
COPY go.mod go.sum Makefile ./
RUN go mod download

# generate & build
ARG VERSION
ARG COMMIT
COPY . .
RUN make deps build

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
COPY --from=build /src/bin/linux-amd64/sver /app/

ENTRYPOINT ["./sver"]

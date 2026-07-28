FROM --platform=${BUILDPLATFORM} golang:1.26-alpine AS builder
LABEL maintainer="Tom Helander <thomas.helander@gmail.com>"

RUN apk add make curl git

WORKDIR /src
COPY .git go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN make GOOS=$TARGETOS GOARCH=$TARGETARCH build

FROM alpine:3.24.1
LABEL maintainer="Tom Helander <thomas.helander@gmail.com>"

WORKDIR /app

RUN apk add curl liquidctl

COPY --from=builder /src/output/liquidctl-exporter .

EXPOSE 9530

ENTRYPOINT ["/app/liquidctl-exporter"]

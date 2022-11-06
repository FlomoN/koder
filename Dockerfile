FROM golang:1.18-alpine as builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS $TARGETOS
ENV GOARCH $TARGETARCH

RUN echo $GOARCH

WORKDIR /app

COPY . .

RUN go mod download && go mod verify

RUN CGO_ENABLED=0 go build -o main -ldflags "-s -w"

FROM gcr.io/distroless/static-debian11

LABEL org.opencontainers.image.source=https://github.com/flomon/koder
COPY --from=builder /app/main /
CMD ["/main"]
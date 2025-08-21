FROM --platform=$BUILDPLATFORM golang:latest as build

WORKDIR /src
COPY * /src

ARG TARGETOS TARGETARCH TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -o rr main.go

FROM scratch

COPY --from=build /src/rr /rr

ENTRYPOINT ["/rr"]

CMD ['']

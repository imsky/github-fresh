FROM golang:1.12-alpine AS build

COPY main.go .

ARG LDFLAGS

RUN GOOS=linux GOARCH=386 go build -ldflags "${LDFLAGS}" -o github-fresh

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/github-fresh /go/github-fresh
ENTRYPOINT ["/go/github-fresh"]

ARG NAME
ARG VERSION
ARG COMMIT
ARG BUILD_DATE

LABEL maintainer="Ivan Malopinsky" org.label-schema.name="${NAME}" org.label-schema.build-date="${BUILD_DATE}" org.label-schema.vcs-ref="${COMMIT}" org.label-schema.version="${VERSION}" org.label-schema.schema-version="1.0"

FROM golang:1.12-alpine AS build

COPY main.go .
RUN go build -o github-fresh

FROM scratch

COPY --from=build /go/github-fresh /go/github-fresh
ENTRYPOINT ["/go/github-fresh"]

ARG NAME
ARG VERSION
ARG COMMIT
ARG BUILD_DATE

LABEL maintainer="Ivan Malopinsky" org.label-schema.name="${NAME}" org.label-schema.build-date="${BUILD_DATE}" org.label-schema.vcs-ref="${COMMIT}" org.label-schema.version="${VERSION}" org.label-schema.schema-version="1.0"

# --------------------------------------------------------------------- dev ---

FROM golang:1.14.7-alpine as dev

RUN apk add --no-cache ca-certificates git gcc libc-dev

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

ARG GOPRIVATE=gitlab.kksharmadevdev.com
ENV GOPRIVATE ${GOPRIVATE}
ARG NETRC_LOCAL

WORKDIR /build

RUN printf "$NETRC_LOCAL" > ~/.netrc

COPY go.mod go.sum ./
RUN go mod download

# Install Compile plug
# Uncomment if manually need to install librdkafka dependency
# RUN apk -U add ca-certificates
# RUN apk update && apk upgrade && apk add pkgconf git bash build-base sudo
# RUN git clone https://github.com/edenhill/librdkafka.git && cd librdkafka && ./configure --prefix /usr && make && make install

# Copy the code into the container
COPY . .

# ------------------------------------------------------------------- build ---

FROM dev as build

ARG VERSION_RELEASE=0.0.0
ENV VERSION_RELEASE $VERSION_RELEASE

ARG VERSION_BUILD=0.0.0-0
ENV VERSION_BUILD $VERSION_BUILD

# Build the application
RUN go build --ldflags "-extldflags -static" -tags musl -o main .

# Move to dist
WORKDIR /dist
RUN cp /build/main .

# ----------------------------------------------------------------- release ---

FROM scratch as release

COPY --from=build /dist/main /

ARG VERSION_RELEASE=0.0.0
ENV VERSION_RELEASE $VERSION_RELEASE

ARG VERSION_BUILD=0.0.0-0
ENV VERSION_BUILD $VERSION_BUILD

ARG KAFKA_SERVER
ENV KAFKA_SERVER $KAFKA_SERVER

ARG KAFKA_TOPIC
ENV KAFKA_TOPIC $KAFKA_TOPIC

ENTRYPOINT [ "/main" ]

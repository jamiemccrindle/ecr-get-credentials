FROM golang:alpine

RUN apk update && apk add ca-certificates git

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY . /go/src/app
RUN go get .
RUN go install .

RUN mkdir -p /opt/build \
    && cp /go/bin/app /opt/build

COPY Dockerfile.run /opt/build/Dockerfile

WORKDIR /opt/build

CMD ["docker", "build", "-t", "ecr-get-credentials", "."]

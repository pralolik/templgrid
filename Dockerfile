FROM golang:1.17.3 AS BUILDER

ARG VERSION
ARG BRANCH
ARG COMMIT

RUN mkdir -p /templgrid
WORKDIR /templgrid
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

ENTRYPOINT ["./templgrid"]
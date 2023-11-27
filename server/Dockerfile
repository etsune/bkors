FROM golang:1.21.0 AS build-stage
WORKDIR /usr/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bkors main.go 

FROM alpine:latest AS build-release-stage 
WORKDIR /
COPY --from=build-stage /bkors /bkors
EXPOSE 1363

ENTRYPOINT ["/bkors"]
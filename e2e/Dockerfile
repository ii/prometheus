FROM golang:1.8
MAINTAINER "t <wolfpack@vulk.coop>"

COPY files/tests /tests
RUN apt-get update && \
    go get -u github.com/prometheus/client_golang/examples/random && \
    cd src/github.com/prometheus/client_golang/examples/random && \
    go get -d && \
    go build && \
    cp -v random /tests/

#./random -listen-address=:8080
#./random -listen-address=:8081
#./random -listen-address=:8082

WORKDIR /tests

CMD ["/bin/bash"]

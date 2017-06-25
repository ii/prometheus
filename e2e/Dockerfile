FROM golang:1.8
MAINTAINER "t <wolfpack@vulk.coop>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    git clone https://github.com/prometheus/client_golang.git && \
    cd client_golang/examples/random && \
    go get -d && \
    go build && \
    mkdir -p /tests && \
    cp -v random /tests/

#./random -listen-address=:8080
#./random -listen-address=:8081
#./random -listen-address=:8082

WORKDIR /tests
COPY files/tests /tests

CMD ["/bin/bash"]

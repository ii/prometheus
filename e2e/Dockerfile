FROM debian:jessie
MAINTAINER "wolfpack <vulk@ii.coop>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl 

COPY tests /tests

CMD ["/bin/bash"]

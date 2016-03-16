FROM golang:1.5.2
RUN cd /tmp
COPY libgit /tmp
RUN cd /tmp && ls -al && tar -xvf libgit2-0.23.4.tar && tar -xvf cmake-2.8.7.tar
RUN cd /tmp/cmake-2.8.7 && ./configure && make && make install
RUN cd /tmp/libgit2-0.23.4/ && cmake . && make && make install
ENV LD_LIBRARY_PATH=/usr/local/lib
RUN apt-get update && apt-get install pkg-config -y --fix-missing
VOLUME /opt/feedhenry/fh-scm/files/
VOLUME /opt/feedhenry/fh-scm/files/backup
RUN mkdir -p /go/src/github.com/fheng/scm-go
COPY . /go/src/github.com/fheng/scm-go
RUN cd /go/src/github.com/fheng/scm-go && go install -v 


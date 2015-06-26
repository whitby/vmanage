FROM golang
ENV CGO_ENABLED=0
WORKDIR /usr/src/go/src
RUN ./make.bash
RUN GOOS=darwin GOARCH=amd64 ./make.bash --no-clean
RUN mkdir -p /go/src/vmanage
RUN go get github.com/go-ldap/ldap
RUN go get github.com/whitby/vcapi
COPY . /go/src/vmanage/
WORKDIR /go/src/vmanage


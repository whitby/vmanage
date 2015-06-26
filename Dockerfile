FROM golang
WORKDIR /usr/src/go/src
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean
RUN GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean
RUN mkdir -p /go/src/vmanage
RUN go get github.com/go-ldap/ldap
RUN go get github.com/whitby/vcapi
COPY . /go/src/vmanage/
WORKDIR /go/src/vmanage


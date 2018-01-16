### STAGE 1: Build ###

# The builder node
FROM golang:latest as builder

# create working directory
WORKDIR /go/src/github.com/made2591/jaker

# copy the content 
COPY . .

# install dependencies
RUN go get ./...

# build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jaker .


### STAGE 2: Setup ###

# The runner node
FROM alpine:latest as runner 

# setup env
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# copy the binary from previous stage
COPY --from=builder /go/src/github.com/made2591/jaker .

COPY crontab /etc/cron.d/checker
RUN chmod 0644 /etc/cron.d/checker
RUN service cron start

# execute
CMD ["./jaker"]
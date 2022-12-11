FROM golang:1.19-alpine AS build

#Install git
RUN apk add --no-cache git

WORKDIR $GOPATH/src/github.com/skeletondev/bones/server

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Build the project and send the output to /bin/bones-server
RUN export GO111MODULE=on && go get . && go build -o /bin/bones-server

FROM golang:1.19-alpine
#Copy the build's output binary from the previous build container
COPY --from=build /bin/bones-server /bin/bones-server
ENTRYPOINT ["/bin/bones-server"]
EXPOSE 8080
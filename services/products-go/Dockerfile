FROM --platform=arm64 golang:1.22.1-alpine

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Bundle app source
COPY . .

RUN go build -o /usr/src/app/bin

ARG PORT=8080

EXPOSE $PORT

CMD [ "/usr/src/app/bin" ]
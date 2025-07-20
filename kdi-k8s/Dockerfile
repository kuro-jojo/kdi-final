# syntax=docker/dockerfile:1

FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV KDI_WORKING_ENV=prod

RUN CGO_ENABLED=1 GOOS=linux go build -o kdi-k8s .

EXPOSE 8090

CMD ["./kdi-k8s"]

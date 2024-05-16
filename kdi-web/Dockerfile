# syntax=docker/dockerfile:1

FROM golang:1.22 

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV KDI_WORKING_ENV=prod

RUN CGO_ENABLED=1 GOOS=linux go build -o kdi-web .

EXPOSE 8080

CMD ["./kdi-web"]
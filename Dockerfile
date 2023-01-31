FROM golang:1.18-alpine as builder
LABEL stage=builder
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o /papaya ./cmd/papaya

FROM alpine

COPY --from=builder /papaya /papaya

EXPOSE 8000

CMD [ "/papaya" ]

FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main ./cmd/papaya
EXPOSE 8000
CMD ["/app/main"]


#############################
## STEP 1 build executable binary
#############################
#
#FROM golang:alpine AS builder
#RUN apk update && apk add --no-cache git
#ADD . /app/
#WORKDIR /app
#RUN go build -o papaya ./cmd/papaya
#
#############################
## STEP 2 build a small image
#############################
#
#FROM scratch
#COPY --from=builder /app/papaya /app/papaya
#RUN chmod 777 /home/papaya/storage
#EXPOSE 8000
#ENTRYPOINT ["/app/papaya"]
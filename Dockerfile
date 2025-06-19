FROM golang:1.24.0

WORKDIR /app
COPY . .
RUN make build
EXPOSE 8000

CMD ["./gotennis"]
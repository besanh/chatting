FROM golang:1.24-bullseye as builder
LABEL stage=Builder

ENV GOPRIVATE=github.com/besanh/chatbot_gpt
ENV GROUP_NAME=chatbot
ENV PROJECT_NAME=chatbot_gpt

ARG GOENV=local

WORKDIR /app
COPY ./api ./api
COPY ./common ./common
COPY ./model ./model
COPY ./pkg ./pkg
COPY ./middleware ./middleware
COPY ./repository ./repository
COPY ./server ./server
COPY ./service ./service
COPY ./config ./config
COPY ./cmd .
COPY ./go.mod .
COPY ./go.sum .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main ./*.go


FROM debian:bullseye-slim as final
LABEL stage=Final
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /root

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yml ./config/config.yml


# Expose port 8000 to the outside world
EXPOSE 8000

#Command to run the executable
CMD ["./main"]

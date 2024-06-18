# Use an official Golang runtime as a parent image
FROM golang:1.21

# Set the working directory inside the container
WORKDIR /app

# Copy the local directory's contents to the container
COPY ./src /app

ENV GOARCH=amd64
ENV GOOS=linux

# Build the Go application
RUN go build -o main

EXPOSE 80

# Command to run the executable
CMD ["./main"]
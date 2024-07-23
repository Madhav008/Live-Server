# Use an official Go runtime as a parent image
FROM golang:latest

# Install FFmpeg
RUN apt-get update && apt-get install -y ffmpeg

# Set the current working directory inside the container
WORKDIR /go/src/app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port 8082
EXPOSE 8082

# Run the Go application
CMD ["./main"]


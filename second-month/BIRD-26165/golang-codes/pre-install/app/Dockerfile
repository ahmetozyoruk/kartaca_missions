# Use an official Golang runtime as a parent image
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . /app


RUN go get ./...
RUN go mod download

# Build the Golang program
RUN go build -o pre-install .

# Run the Golang application
CMD ["./pre-install"]
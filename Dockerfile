# Use official Go image
FROM golang:1.25

# Set working directory, don't forget to change the project name
WORKDIR /golang_template_v3

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy project files
COPY . .

# Build the app, don't forget to change the project name
RUN go build -o golang_template_v3
# Run the binary
CMD ["./golang_template_v3"]
# Build the CLI application
build:
	go build -o bin/cli main.go

# Add .PHONY to declare build as a phony target
.PHONY: build

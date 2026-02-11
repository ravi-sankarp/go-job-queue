.PHONY: all test mock-server air dev
all:
	go run main.go
test:
	go test
mock-server:
	@echo "Starting Mock Server"
	go run mock-server/main.go
air:
	air
# dev server running both mock-server and scheduler and reloads on change
dev:
	go run mock-server/main.go & air

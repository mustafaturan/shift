all: test_race test coverage
test_race:
	go test ./... -race
test:
	go test ./... -coverprofile=coverage.out
coverage:
	go tool cover -html=coverage.out

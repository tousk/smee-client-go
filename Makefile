smee: sse.go main.go
	go get -d
	go build -ldflags="-s -w" -o smee
clean:
	rm -f go

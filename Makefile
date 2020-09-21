run:
	cd server && \
	go run server.go &
	cd client && \
	go test client_test.go -v
build:
	cd cmd/app; \
	go get; \
	cd ../../; \
	go build -o syncfolder cmd/app/main.go;


bench:
	cd test; \
	go test -bench=BenchmarkCopyFile -benchmem -benchtime=10s

tests:
	cd test; \
	go test ./... -v -coverpkg=../../...

cover:
	cd test; \
	go test ./... -v -coverpkg=../../... -coverprofile=cover.out; \
	go tool cover -html=cover.out
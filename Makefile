test:
	go test --race --count=1 -coverprofile coverage.out $$(go list ./... | grep -v github.com/DanLavine/goasync/internal/examples | grep -v github.com/DanLavine/goasync/goasyncfakes)
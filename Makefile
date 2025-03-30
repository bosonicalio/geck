# This follows the convention of the package name according to big tech companies style guides.
gen-mock:
	mockgen -source=${PACKAGE}/${FILE}.go -destination=${PACKAGE}mock/${FILE}.go -package=${PACKAGE}mock -typed

run-all-tests:
	go test ./... -tags=integration,e2e -v -cover -race -p=10

all:
	mkdir -p build
	go build -o build/SQLUpdate github.com/Catofes/ipfscdn/manager/sql/update
	go build -o build/node github.com/Catofes/ipfscdn/node/run

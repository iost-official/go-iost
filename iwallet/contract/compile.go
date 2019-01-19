package contract

//go:generate npm install
//go:generate ./node_modules/.bin/webpack
//go:generate go run gen.go
//go:generate rm -rf ./dist && ./node_modules

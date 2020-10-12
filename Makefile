.SILENT: main
.PHONY: main run test
.DEFAULT_GOAL := main
main: # don't change this line; first line is the default target in make <= 3.79 despite .DEFAULT_GOAL
	echo "market commands:"
	echo "run, test"
run:
	go run -trimpath cmd/market/*.go serve -config cmd/market/local.json
test:
	./scripts/test.sh

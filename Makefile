all:
	cd base; go install -v ./...
	cd cases; go install -v ./...

install: all
	@echo


clean:
	@echo

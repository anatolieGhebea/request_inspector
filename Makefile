build_release:
	@echo "Building for Linux"
	GOOS=linux GOARCH=amd64 go build -o ./dist/linux/$(V)/request_inspector 
	cp -r ./static ./dist/linux/$(V)/static
	cp ./Makefile ./dist/linux/$(V)/Makefile
	tar -czf ./dist/linux/request_inspector_v$(V).tar.gz -C ./dist/linux/$(V)/ .
	rm -rf ./dist/linux/$(V)

install_dev:
	@echo "Installing locally"
	@if [ -z "$(f)" ]; then \
		echo "Error: 'f' variable is not set. Please specify the file to install. "; \
		echo "USAGE: make install_local f=request_inspector_v<version_number>.tar.gz"; \
		echo "run ls -la ./dist/linux to see the available files for install"; \
		exit 1; \
	fi
	@echo "Installing locally"
	tar -xzvf ./dist/linux/$(f) -C ./bin
	chmod +x ./bin/request_inspector

install_local:
	@echo "Installing locally"
	@if [ -z "$(f)" ]; then \
		echo "Error: 'f' variable is not set. Please specify the file to install. "; \
		echo "USAGE: make install_local f=request_inspector_v<version_number>.tar.gz"; \
		echo "run ls -la ./dist/linux to see the available files for install"; \
		exit 1; \
	fi
	@echo "Installing locally"
	tar -xzvf ./$(f) -C ./bin
	chmod +x ./bin/request_inspector

start_local:
	@echo "Starting locally"
	./bin/request_inspector &

stop_local:
	@echo "Stopping locally"
	killall request_inspector

show_running:
	@echo "Running processes"
	ps aux | grep request_inspector
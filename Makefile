build_linux:
	@echo "Building for Linux"
	GOOS=linux GOARCH=amd64 go build -o ./dist/linux/$(V)/request_inspector 
	tar -czf ./dist/linux/request_inspector_v$(V).tar.gz -C ./dist/linux/$(V)/ .
	rm -rf ./dist/linux/$(V)

install_local:
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

start_local:
	@echo "Starting locally"
	./bin/request_inspector &

stop_local:
	@echo "Stopping locally"
	killall request_inspector

show_running:
	@echo "Running processes"
	ps aux | grep request_inspector

install_system:
	@echo "TODO: Installing as system service"

start_system:
	@echo "TODO: Starting as system service"

stop_system:
	@echo "TODO: Stopping as system service"

restart_system:
	@echo "TODO: Restarting as system service"

enable_system:
	@echo "TODO: Enabling as system service"

disable_system:
	@echo "TODO: Disabling as system service"
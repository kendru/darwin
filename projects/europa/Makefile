.PHONY: build

build: target/main.wasm

# target/main.wasm:
# 	zig build-lib src/main.zig -prefix target -target wasm32-freestanding -dynamic -O ReleaseSmall

target/main.wasm:
	zig build-lib src/main.zig -target wasm32-freestanding -dynamic

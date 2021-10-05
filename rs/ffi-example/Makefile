.PHONY: build

CC=clang
DEPS = point.h

lib%.so: %.c $(DEPS)
	$(CC) --shared -Wall -o $@ $<

build: libpoint.so target/release/ffi-example

target/release/ffi-example: src/main.rs
	LIBRARY_PATH=".:$$LIBRARY_PATH" cargo build --release


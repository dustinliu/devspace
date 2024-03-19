app := devspace

export TC_RUNTIME_ENV=dev

build:
	@cargo build

release-build:
	@cargo build --release

test:
	@RUST_BACKTRACE=full cargo test

clean:
	@cargo clean

.PHONY: build release-buil clean test

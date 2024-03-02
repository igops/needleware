protoc-go:
	docker run --rm \
		-v $$PWD:/defs \
		namely/protoc-all \
		-f proto/needleware.proto \
		-l go
	rm -rf ./pb
	mv ./gen/pb-go/* .
	rm -rf gen

.PHONY: schema
schema:
	protoc --go_out=paths=source_relative:. test.proto

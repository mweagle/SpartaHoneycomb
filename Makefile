.DEFAULT_GOAL=build
.PHONY: build test provision describe

clean:
	go clean .

test: build
	go test -v .

delete:
	go run main.go delete

explore:
	go run main.go --level info explore

provision:
	go run main.go provision --level info --s3Bucket $(S3_BUCKET)

describe: build
	go run main.go --level info describe --out ./graph.html --s3Bucket $(S3_BUCKET)

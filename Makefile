build:
	export GO111MODULE=on
	cp insta/main.go insta/main.bak
	cd insta; go mod tidy; cd ..
	env GOOS=linux go build -ldflags="-s -w" -o bin/insta insta/main.go

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: deploy
deploy: clean build
	sls deploy --verbose --region eu-central-1 --aws-profile "suka.yoga"

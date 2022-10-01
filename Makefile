REPO = github.com/findbed/app
CWD = /go/src/$(REPO)
IMG = registry.$(REPO)
TAG = latest

test: lint build

lint:
	@-docker run --rm -t -v $(CURDIR):$(CWD) -w $(CWD) \
		-v $$HOME/.docker-go-cache:/go/pkg \
		-e GOFLAGS=-mod=mod \
		golangci/golangci-lint:latest-alpine sh -c '\
			golangci-lint run ${LINT_FLAGS}'

build: buildfs
	@docker build -t $(IMG):$(TAG) .

buildfs:
	@docker run --rm \
		-e CWD=$(CWD) \
		-v $(CURDIR)/runner:/runner \
		-v $(CURDIR)/buildfs:/build \
		-v $(CURDIR):$(CWD) \
		-v $$HOME/.docker-go-cache:/root/go/pkg \
		-e TAG=$(TAG) \
		-e GOFLAGS=-mod=mod \
		imega/base-builder:1.9.9 \
		--packages=" \
			libcrypto3@main \
			libssl3@main \
			xh@edge-community \
			ca-certificates \
			busybox" \
		-d=" \
			openssh-client-default \
			tzdata \
			curl \
			git \
			go@community \
			gcc \
			alpine-sdk \
		"

clean:
	@-GO_IMG=$(GO_IMG) CWD=$(CWD) docker-compose rm -sfv
	@-docker run --rm -v $(CURDIR):/data -w /data alpine rm -rf buildfs

release: test

zinc:
	mkdir -p data
	docker run -d -v $(CURDIR):/data -e ZINC_DATA_PATH="/data" -p 4080:4080 \
		-e ZINC_FIRST_ADMIN_USER=admin \
		-e ZINC_FIRST_ADMIN_PASSWORD=Complexpass#123 \
		--name zinc \
		--network zinc_default \
		public.ecr.aws/zinclabs/zinc:latest

index:
	docker run -t -v $(CURDIR):/data -w /data \
		--network zinc_default alpine sh -c '\
		apk add curl && \
		curl -X POST http://zinc:4080/api/index -u admin:Complexpass#123 -H "Content-Type: application/json" -d @create_index.json'

upd:
	docker run -t -v $(CURDIR):/data -w /data \
		--network zinc_default alpine sh -c '\
		apk add curl && \
		curl -X PUT http://zinc:4080/api/thing/_mapping -u admin:Complexpass#123 -H "Content-Type: application/json" -d @update_index.json'


doc:
	docker run -t -v $(CURDIR):/data -w /data \
		--network zinc_default alpine sh -c '\
		apk add curl && \
		curl -X POST http://zinc:4080/api/thing/_doc -u admin:Complexpass#123 -H "Content-Type: application/json" -d @create_doc.json'

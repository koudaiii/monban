REVISION  := $(shell git rev-parse --short HEAD)
VERSION := v0.1.0

.PHONY: build
build: dep-ensure
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook .
	docker build --no-cache -t quay.io/koudaiii/monban:$(REVISION) .
	rm -rf webhook

.PHONY: push
push:
	docker push quay.io/koudaiii/monban:$(REVISION)

.PHONY: release
release:
	docker tag quay.io/koudaiii/monban:$(REVISION) quay.io/koudaiii/monban:$(VERSION)
	docker push quay.io/koudaiii/monban:$(VERSION)

.PHONY: clean
clean:
	rm -rf vendor/*
	rm -f certs/dst/*.*

.PHONY: dep-ensure
dep-ensure:
	dep version || go get -u github.com/golang/dep/cmd/dep
	dep ensure -v
	dep prune -v
	find vendor -name '*_test.go' -delete

.PHONY: ssl-generate
ssl-generate:
	cfssl gencert -initca certs/ca-csr.json | cfssljson -bare certs/dst/ca
	cfssl gencert \
	  -ca=certs/dst/ca.pem \
	  -ca-key=certs/dst/ca-key.pem \
	  -config=certs/cert-config.json \
	  -profile=default \
	  certs/admission-webhook-csr.json | cfssljson -bare certs/dst/admission-webhook

.PHONY: deploy
deploy: ssl-generate
	cat kubernetes/validating-webhook-configuration.yaml | sed "s/__CA__/$(shell cat certs/dst/ca.pem | base64)/" | kubectl apply -f -
	cat kubernetes/monban.yaml | sed "s/__CA__/$(shell cat certs/dst/ca.pem | base64)/"  | sed "s/__TLS_CERT__/$(shell cat certs/dst/admission-webhook.pem | base64)/" | sed "s/__TLS_KEY__/$(shell cat certs/dst/admission-webhook-key.pem | base64)/" | kubectl apply -f -

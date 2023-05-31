.PHONY: run clean

run: build/domain.crt
	go run main.go build/domain.crt build/domain.key

clean: | build
	-@rm build/**

build:
	mkdir -p $@

build/ca_password: | build
	openssl rand -hex 16 > $@

build/rootCA.key: build/ca_password
	openssl req -x509 -sha256 -days 1825 \
	  -newkey rsa:2048 \
	  -passout file:build/ca_password \
	  -subj "/CN=localhost" \
	  -keyout build/rootCA.key -out build/rootCA.crt

build/domain.csr: build/rootCA.key
	openssl req -newkey rsa:2048 -keyout build/domain.key \
	-nodes \
	-passin file:build/ca_password \
	-subj "/CN=localhost" \
	-out build/domain.csr

build/domain.crt: build/domain.csr
	openssl x509 -req -CA build/rootCA.crt -CAkey build/rootCA.key \
		-in build/domain.csr -out build/domain.crt -days 365 \
		-passin file:build/ca_password \
		-CAcreateserial -extfile cert.ext

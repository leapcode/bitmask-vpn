IMAGE=bitmask-vpn
REGISTRY=registry.0xacab.org/leap/docker

build:
	docker build -t $(REGISTRY)/$(IMAGE):latest .
push:
	docker push $(REGISTRY)/$(IMAGE):latest
prune:
	docker system prune
login:
	docker login $(REGISTRY)
shell:
	docker run -v `python3 -c 'import os; print(os.path.abspath(".."))'`:/bitmask-vpn.orig -it --rm $(REGISTRY)/$(IMAGE):latest  /bin/bash

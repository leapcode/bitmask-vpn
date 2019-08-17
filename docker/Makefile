IMAGE=bitmask-vpn
REGISTRY=0xacab.org:4567/leap/docker

build:
	docker build -t $(REGISTRY)/$(IMAGE):latest .
push:
	docker push $(REGISTRY)/$(IMAGE):latest
prune:
	docker system prune
login:
	docker login 0xacab.org:4567
shell:
	docker run -it --rm $(REGISTRY)/$(IMAGE):latest  /bin/bash

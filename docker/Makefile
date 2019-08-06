IMAGE=bitmask-vpn

build:
	docker build -t 0xacab.org:4567/leap/docker/$(IMAGE):latest .
push:
	docker push 0xacab.org:4567/leap/docker/$(IMAGE):latest
prune:
	docker system prune
login:
	docker login 0xacab.org:4567

IMAGE=bitmask-vpn
REGISTRY=registry.0xacab.org/leap/docker

build:
	mkdir -p mods && cp ../go.mod ../go.sum mods/ && cp -r ../packages mods/
	docker build -t $(REGISTRY)/$(IMAGE):latest .
push:
	docker push $(REGISTRY)/$(IMAGE):latest
prune:
	docker system prune
login:
	docker login $(REGISTRY)
package_win:
	docker run \
	    -e XBUILD=win \
	    -e PROVIDER=riseup \
	    -v `python3 -c 'import os; print(os.path.abspath(".."))'`:/bitmask-vpn.host \
	    --cpus="4" \
	    -it --rm \
	    $(REGISTRY)/$(IMAGE):latest  \
	    /builder.sh
shell:
	docker run -v `python3 -c 'import os; print(os.path.abspath(".."))'`:/bitmask-vpn.host -it --rm $(REGISTRY)/$(IMAGE):latest  /bin/bash

#########################################################################
# Docker build environment for BitmaskVPN
# (c) LEAP Encryption Access Project, 2019-2020
#########################################################################

IMAGE=bitmask-vpn
REGISTRY=registry.0xacab.org/leap/docker

PROVIDER=riseup
PRODUCT=$(shell echo $(PROVIDER) | sed -e "s/\b\(.\)/\u\1/g")VPN

UNINSTALLER_NSIS="../deploy/produce-bitmask-uninstaller.exe"
UNINSTALLER_WINE=".wine/drive_c/bitmask-uninstall.exe"
UNINSTALLER_STAGING="../build/windows/staging/uninstall.exe"
UNINSTALLER_SIGNED="../build/windows/staging/uninstall-signed.exe"
SIGNCODE=osslsigncode
SECRETS="leap/secrets/windoze"

build:
	mkdir -p mods && cp ../go.mod ../go.sum mods/ && cp -r ../packages mods/
	docker build -t $(REGISTRY)/$(IMAGE):latest .
deps:
	sudo apt install osslsigncode
push:
	docker push $(REGISTRY)/$(IMAGE):latest
prune:
	docker system prune
login:
	docker login $(REGISTRY)

package_win: clean_win package_win_stage_1 sign_win_uninstaller_in_host package_win_stage_2 sign_win_installer_in_host

clean_win:
	@rm -f ../deploy/*.exe

package_win_stage_1:
	docker run \
	    -e PROVIDER=$(PROVIDER) \
	    -e XBUILD=win \
	    -e STAGE=1 \
	    -v `python3 -c 'import os; print(os.path.abspath(".."))'`:/bitmask-vpn.host \
	    --cpus="4" \
	    -it --rm \
	    $(REGISTRY)/$(IMAGE):latest  \
	    /builder.sh

sign_win_uninstaller_in_host:
	@rm -f $(UNINSTALLER_STAGING)
	@rm -f $(UNINSTALLER_WINE)
	@echo "[+] running nsis uninstaller under wine..."
	@wine $(UNINSTALLER_NSIS) || echo "[+] wine finished (it exits badly)"
	@cp ~/$(UNINSTALLER_WINE) $(UNINSTALLER_STAGING) && echo "[+] copied uninstaller to staging dir:" $(UNINSTALLER_STAGING)
	@$(SIGNCODE) sign \
	    -certs ~/$(SECRETS)/win.crt \
	    -key ~/$(SECRETS)/key.pem \
	    -pass `gpg --decrypt ~/$(SECRETS)/../windoze.gpg` \
	    -n "$(PRODUCT) Uninstaller" -i "https://leap.se/" \
	    -in $(UNINSTALLER_STAGING) \
	    -out $(UNINSTALLER_SIGNED)
	@echo "[+] signed uninstaller in:" $(UNINSTALLER_SIGNED)

package_win_stage_2:
	docker run \
	    -e PROVIDER=$(PROVIDER) \
	    -e XBUILD=win \
	    -e STAGE=2 \
	    -v `python3 -c 'import os; print(os.path.abspath(".."))'`:/bitmask-vpn.host \
	    --cpus="4" \
	    -it --rm \
	    $(REGISTRY)/$(IMAGE):latest  \
	    /builder.sh

sign_win_installer_in_host:
	$(eval INSTALLER := $(shell ls -tr ../deploy/$(PRODUCT)*.exe | tail -n 1))
	$(eval INSTALLER_SIGNED := $(shell echo $(INSTALLER) | sed 's/\.exe/-signed.exe/'))
	@$(SIGNCODE) sign \
	    -certs ~/$(SECRETS)/win.crt \
	    -key ~/$(SECRETS)/key.pem \
	    -pass `gpg --decrypt ~/$(SECRETS)/../windoze.gpg` \
	    -n "$(PRODUCT) Installer" -i "https://leap.se/" \
	    -in $(INSTALLER) \
	    -out ../deploy/$(INSTALLER_SIGNED)
	@echo "[+] signed installer in:" $(INSTALLER_SIGNED)

shell:
	docker run -v `python3 -c 'import os; print(os.path.abspath(".."))'`:/bitmask-vpn.host -it --rm $(REGISTRY)/$(IMAGE):latest  /bin/bash

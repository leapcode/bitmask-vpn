build_static_openvpn:
	pkg/thirdparty/openvpn/build_openvpn.sh

upload_openvpn:
	rsync --rsh='ssh' -avztlpog --progress --partial ~/openvpn_build/sbin/openvpn* downloads.leap.se:./public/thirdparty/linux/openvpn/

download_openvpn:
	wget https://downloads.leap.se/thirdparty/linux/openvpn/openvpn

clean_openvpn_build:
	rm -rf ~/openvpn_build

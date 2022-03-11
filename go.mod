module 0xacab.org/leap/bitmask-vpn

go 1.17

require (
	0xacab.org/leap/obfsvpn v0.0.0-20220311174134-724b17ec5b25
	git.torproject.org/pluggable-transports/goptlib.git v1.2.0
	git.torproject.org/pluggable-transports/snowflake.git v1.1.0
	github.com/ProtonMail/go-autostart v0.0.0-20181114175602-c5272053443a
	github.com/apparentlymart/go-openvpn-mgmt v0.0.0-20161009010951-9a305aecd7f2
	github.com/cretz/bine v0.2.0
	github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19
	github.com/pion/webrtc/v3 v3.0.15
	github.com/sevlyar/go-daemon v0.1.5
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/kcp-go/v5 v5.6.1
	github.com/xtaci/smux v1.5.15
	golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5
)

require (
	github.com/dchest/siphash v1.2.2 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/reedsolomon v1.9.9 // indirect
	github.com/mmcloughlin/avo v0.0.0-20200803215136-443f81d77104 // indirect
	github.com/pion/datachannel v1.4.21 // indirect
	github.com/pion/dtls/v2 v2.0.8 // indirect
	github.com/pion/ice/v2 v2.0.15 // indirect
	github.com/pion/interceptor v0.0.10 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.4 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.6 // indirect
	github.com/pion/rtp v1.6.2 // indirect
	github.com/pion/sctp v1.7.11 // indirect
	github.com/pion/sdp/v3 v3.0.4 // indirect
	github.com/pion/srtp/v2 v2.0.2 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.12.3 // indirect
	github.com/pion/turn/v2 v2.0.5 // indirect
	github.com/pion/udp v0.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/smartystreets/assertions v0.0.0-20180927180507-b2de0cb4f26d // indirect
	github.com/templexxx/cpu v0.0.7 // indirect
	github.com/templexxx/xorsimd v0.4.1 // indirect
	github.com/tjfoc/gmsm v1.3.2 // indirect
	gitlab.com/yawning/obfs4.git v0.0.0-20220204003609-77af0cba934d // indirect
	golang.org/x/crypto v0.0.0-20220307211146-efcb8507fb70 // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/tools v0.0.0-20200808161706-5bf02b21f123 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

// The changes to obfs4 in the next commit (393aca8) are not backwards
// compatible, contrary to what the documentation says. Temporarily use an older
// version until the gateways are updated.
replace gitlab.com/yawning/obfs4.git => gitlab.com/yawning/obfs4.git v0.0.0-20210511220700-e330d1b7024b

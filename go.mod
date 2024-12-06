module 0xacab.org/leap/bitmask-vpn

go 1.22.2

require (
	0xacab.org/leap/obfsvpn v1.3.1-0.20241121155258-e6b06efc4456
	git.torproject.org/pluggable-transports/goptlib.git v1.3.0
	git.torproject.org/pluggable-transports/snowflake.git v1.1.0
	github.com/ProtonMail/go-autostart v0.0.0-20210130080809-00ed301c8e9a
	github.com/cretz/bine v0.2.0
	github.com/dchest/siphash v1.2.3 // indirect
	github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19
	github.com/pion/webrtc/v3 v3.2.44
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/kcp-go/v5 v5.6.11
	github.com/xtaci/smux v1.5.24
	// Do not update obfs4 past e330d1b7024b, a backwards incompatible change was
	// made that will break negotiation!! riseup should move to the newest asap.
	gitlab.com/yawning/obfs4.git v0.0.0-20231012084234-c3e2d44b1033 // indirect
	golang.org/x/sys v0.23.0
)

require (
	0xacab.org/leap/bitmask-core v0.0.0-20241015185856-0812b9aadf98
	github.com/natefinch/npipe v0.0.0-20160621034901-c1b8fa8bdcce
	github.com/prometheus-community/pro-bing v0.4.0
	github.com/rs/zerolog v1.33.0
	github.com/stretchr/testify v1.9.0
)

require (
	0xacab.org/leap/tunnel-telemetry v0.0.0-20240830081933-7328bb50078b // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/asdine/storm/v3 v3.2.1 // indirect
	github.com/babolivier/go-doh-client v0.0.0-20201028162107-a76cff4cb8b6 // indirect
	github.com/cloudflare/circl v1.3.9 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/runtime v0.28.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/strfmt v0.23.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20231212022811-ec68065c825e // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/klauspost/reedsolomon v1.12.2 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/onsi/ginkgo/v2 v2.13.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pion/datachannel v1.5.7 // indirect
	github.com/pion/dtls/v2 v2.2.11 // indirect
	github.com/pion/ice/v2 v2.3.28 // indirect
	github.com/pion/interceptor v0.1.29 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.12 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.14 // indirect
	github.com/pion/rtp v1.8.6 // indirect
	github.com/pion/sctp v1.8.18 // indirect
	github.com/pion/sdp/v3 v3.0.9 // indirect
	github.com/pion/srtp/v2 v2.0.18 // indirect
	github.com/pion/stun v0.6.1 // indirect
	github.com/pion/transport/v2 v2.2.5 // indirect
	github.com/pion/turn/v2 v2.1.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/quic-go/quic-go v0.48.2 // indirect
	github.com/refraction-networking/utls v1.6.7 // indirect
	github.com/smartystreets/assertions v0.0.0-20180927180507-b2de0cb4f26d // indirect
	github.com/templexxx/cpu v0.1.0 // indirect
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b // indirect
	github.com/templexxx/xorsimd v0.4.2 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/xtaci/kcp-go v5.4.20+incompatible // indirect
	gitlab.com/yawning/edwards25519-extra v0.0.0-20231005122941-2149dcafc266 // indirect
	gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib v1.5.0 // indirect
	go.etcd.io/bbolt v1.3.10 // indirect
	go.mongodb.org/mongo-driver v1.16.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

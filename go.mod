module github.com/LalatinaHub/LatinaBot

go 1.21

toolchain go1.21.0

require (
	github.com/LalatinaHub/LatinaApi v0.0.0-20230826002457-0575c2547ddb
	github.com/LalatinaHub/LatinaSub-go v0.0.0-20230826004229-80a33a5abb43
	github.com/NicoNex/echotron/v3 v3.26.0
)

replace (
	github.com/LalatinaHub/LatinaApi => ../LatinaApi
	github.com/LalatinaHub/LatinaSub-go => ../LatinaSub-go
	github.com/sagernet/sing-box => github.com/LalatinaHub/sing-box v0.0.0-20230825233219-30bbdbbc48f0
)

require (
	berty.tech/go-libtor v1.0.385 // indirect
	github.com/Dreamacro/clash v1.18.0 // indirect
	github.com/Dreamacro/protobytes v0.0.0-20230617041236-6500a9f4f158 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/bytedance/sonic v1.10.0 // indirect
	github.com/caddyserver/certmagic v0.19.2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/cretz/bine v0.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.9.1 // indirect
	github.com/go-chi/chi/v5 v5.0.10 // indirect
	github.com/go-chi/cors v1.2.1 // indirect
	github.com/go-chi/render v1.0.3 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.15.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gofrs/uuid/v5 v5.0.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/pprof v0.0.0-20230821062121-407c9e7a662f // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/insomniacslk/dhcp v0.0.0-20230816195147-b3ca2534940d // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/libdns/libdns v0.2.1 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mholt/acmez v1.2.0 // indirect
	github.com/miekg/dns v1.1.55 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/onsi/ginkgo/v2 v2.12.0 // indirect
	github.com/ooni/go-libtor v1.1.8 // indirect
	github.com/oschwald/maxminddb-golang v1.12.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.3.3 // indirect
	github.com/sagernet/cloudflare-tls v0.0.0-20221031050923-d70792f4c3a0 // indirect
	github.com/sagernet/go-tun2socks v1.16.12-0.20220818015926-16cb67876a61 // indirect
	github.com/sagernet/gvisor v0.0.0-20230808113425-d8f9f5e110c4 // indirect
	github.com/sagernet/netlink v0.0.0-20220905062125-8043b4a9aa97 // indirect
	github.com/sagernet/quic-go v0.0.0-20230825040534-0cd917b2ddda // indirect
	github.com/sagernet/reality v0.0.0-20230406110435-ee17307e7691 // indirect
	github.com/sagernet/sing v0.2.10-0.20230824115837-8d731e68853a // indirect
	github.com/sagernet/sing-box v1.4.0-rc.3 // indirect
	github.com/sagernet/sing-dns v0.1.9-0.20230824120133-4d5cbceb40c1 // indirect
	github.com/sagernet/sing-mux v0.1.3-0.20230811111955-dc1639b5204c // indirect
	github.com/sagernet/sing-shadowsocks v0.2.4 // indirect
	github.com/sagernet/sing-shadowsocks2 v0.1.3 // indirect
	github.com/sagernet/sing-shadowtls v0.1.4 // indirect
	github.com/sagernet/sing-tun v0.1.12-0.20230821065522-7545dc2d5641 // indirect
	github.com/sagernet/sing-vmess v0.1.7 // indirect
	github.com/sagernet/smux v0.0.0-20230312102458-337ec2a5af37 // indirect
	github.com/sagernet/tfo-go v0.0.0-20230816093905-5a5c285d44a6 // indirect
	github.com/sagernet/utls v0.0.0-20230309024959-6732c2ab36f2 // indirect
	github.com/sagernet/websocket v0.0.0-20220913015213-615516348b4e // indirect
	github.com/sagernet/wireguard-go v0.0.0-20230807125731-5d4a7ef2dc5f // indirect
	github.com/scjalliance/comshim v0.0.0-20230315213746-5e51f40bd3b9 // indirect
	github.com/sethvargo/go-password v0.2.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/u-root/uio v0.0.0-20230305220412-3e8cd9d6bf63 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.25.0 // indirect
	go4.org/netipx v0.0.0-20230824141953-6213f710f925 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.12.1-0.20230815132531-74c255bcf846 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
)

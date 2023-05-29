module autobangumi-go

go 1.20

require (
	github.com/PuerkitoBio/goquery v1.8.1
	github.com/anacrolix/torrent v1.50.0
	github.com/antlabs/strsim v0.0.3
	github.com/cyruzin/golang-tmdb v1.5.0
	github.com/dustin/go-humanize v1.0.1
	github.com/fsnotify/fsnotify v1.5.4
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/lyqingye/pikpak-go v0.0.0-20230529054323-cbdaca3a3b8a
	github.com/nssteinbrenner/anitogo v0.0.0-20200907113149-eb04a0056b4a
	github.com/rs/zerolog v1.29.1
	github.com/siku2/arigo v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.2
	github.com/tendermint/tm-db v0.6.7

)

require (
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/anacrolix/missinggo v1.3.0 // indirect
	github.com/anacrolix/missinggo/v2 v2.7.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/bradfitz/iter v0.0.0-20191230175014-e8f45d346db8 // indirect
	github.com/cenkalti/hub v1.0.1 // indirect
	github.com/cenkalti/rpc2 v0.0.0-20210604223624-c1acbc6ec984 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cosmos/gorocksdb v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2 // indirect
	github.com/dgraph-io/ristretto v0.0.3-0.20200630154024-f66de99634de // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace github.com/lyqingye/pikpak-go => /home/lyqingye/workspace/pikpak-go/
replace github.com/siku2/arigo => github.com/lyqingye/arigo v0.0.0-20230527062939-ade95127cd9e

replace github.com/nssteinbrenner/anitogo => github.com/lyqingye/anitogo v0.0.0-20230517021436-770746b18c27

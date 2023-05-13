module pikpak-bot

go 1.20

require (
	github.com/dustin/go-humanize v1.0.1
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/lyqingye/pikpak-go v0.0.0-20230507065106-d0a30cd939f9
	github.com/rs/zerolog v1.29.1
	github.com/siku2/arigo v0.2.0
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.2
)

require (
	github.com/cenkalti/hub v1.0.1 // indirect
	github.com/cenkalti/rpc2 v0.0.0-20210604223624-c1acbc6ec984 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace github.com/lyqingye/pikpak-go => /home/lyqingye/workspace/pikpak-go/
replace github.com/siku2/arigo => github.com/lyqingye/arigo v0.0.0-20230506162102-f6574a6e57c5

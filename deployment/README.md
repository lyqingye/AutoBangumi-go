## 简述
1. 本项目需要访问 `tmdb` `mikan` 等一些国外网站，代理问题请自行解决.  
2. `tmdb`需要用到获取免费的api token，请自行去申请。https://developer.themoviedb.org/v4/docs  
3. 由于没有`webui` ，所以使用了`telegram bot` 进行交互，所以自行创建`bot`.  
4. 由于目前`pikpak`可以免费账号每天下载5次，而且有6gb空间， 所以可以自行申请多个邮箱账号。这样就可以白嫖。

## 依赖环境
- docker
- docker-compose
- make

## 配置说明
```toml
[DB]
Host = "db" # 数据库连接host
Port = "5432" # 数据库端口
User = "lyqingye" 
Password = "lyqingye"
Name = "auto_bangumi" # 数据库db名称
MaxConns = 20 # 数据库最大连接数
LogLevel = "info"

[QB]
Enable = false # 是否启动QB下载器，如果启用了，那么当pikpak无法下载时，将会使用QB下载
Endpoint = "http://qb:8888" # QB API
Username = "admin" # QB 用户名
Password = "adminadmin" # QB密码
DownloadDir = "/downloads" # 下载目录

[Aria2]
WsUrl = "ws://aria2:6800/jsonrpc" # aria2 websocket url
Secret = "123456" # aria2 rpc 密钥
DownloadDir = "/downloads" # 下载目录

[Pikpak]
OfflineDownloadTimeout = "2m" # 使用pikpak 离线下载超时时间

[Cache]
CacheDir = "/cache" # 缓存目录，用户缓存http请求
ClearCacheInterval = "24h" # 缓存清理周期，24h清理一次

[AutoBangumi]
BangumiCompleteInterval = "60m" # 番剧刷新周期，会自动补全确实集数， 这个周期可以设置得长一点
RssRefreshInterval = "5m" # 刷新订阅的周期， 这个周期可以设置的快一点，通常订阅都是最新的内容。（暂不支持）

[TMDB]
Token = "" # TMDB 的token , 自行去申请 https://developer.themoviedb.org/v4/docs

[BangumiTV]
Endpoint = "https://api.bgm.tv/v0" # bangumi TV的 API 域名

[Mikan]
Endpoint = "https://mikanani.me" # mikan的域名

[TelegramBot]
Enable = false # 是否启用 telegram bot
Token = "" # telegram bot 的token
```

## bot 命令
- 添加pikpak账号
```shell
/add_pikpak_account 账号 密码
```
- 追番
```shell
/search_bangumi 因想当冒险者而前往大都市的女儿已经升到了S级
```
## 本地部署
```shell
# 本机部署所有服务
make local-deploy

# 停止并且清理缓存数据
make clean-local-deploy

# 停止，但不清理数据
make stop-local-deploy
```

## 相关服务部署情况
- postgres
- aria2
- aria2 webUI
```bash
RPC端口： 6880
RPCSecret: 123456
```
- qbittorrent
```shell
WEBUI： 8888
用户名： admin
密码： adminadmin
```
- auto-bangumi

**注： 容器数据和下载的番剧数据都放在 `deployment` 目录下**
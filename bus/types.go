package bus

// Topics
const (
	RSSTopic   = "topic-tss"
	Aria2Topic = "topic-aria2"
	QBTopic    = "topic-qb"
)

// RSS Event types
const (
	RSSUpdateEventType      = "event-rss-update"
	RSSSubscribeEventType   = "event-rss-subcribe"
	RSSUnSubscribeEventType = "event-rss-unsubscribe"
	RSSParseErrEventType    = "event-rss-parse-err"
)

type RSSUpdateEvent struct {
}

// Aria2 Event types
const (
	Aria2AddTaskEventType      = "event-aria2-add-task"
	Aria2TaskCompleteEventType = "event-aria2-task-complete"
)

// QBittorrent Event types
const (
	QBAddTorrentEventType       = "event-qb-add-torrent"
	QBDownloadCompleteEventType = "event-qb-download-complete"
)

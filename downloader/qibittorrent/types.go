package qibittorrent

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

var (
	ErrTorrentNotFound  = errors.New("torrent not found")
	ErrMissingParameter = errors.New("missing parameter")
	ErrPath             = errors.New("invalid newPath or oldPath, or newPath already in use")
)

type AddTorrentOptions struct {
	Urls               []string
	Torrents           []byte
	SavePath           string
	Cookie             string
	Category           string
	Tags               []string
	SkipChecking       bool
	Paused             bool
	RootFolder         bool
	Rename             string
	UpLimit            int
	DlLimit            int
	RatioLimit         float64
	SeedingLimit       int
	AutoTMM            bool
	SequentialDownload bool
	FirstLastPiecePrio bool
}

func (options *AddTorrentOptions) toMap() map[string]string {
	return map[string]string{
		"urls":               strings.Join(options.Urls, "\n"),
		"torrents":           hex.EncodeToString(options.Torrents),
		"savepath":           options.SavePath,
		"cookie":             options.Cookie,
		"category":           options.Category,
		"tags":               strings.Join(options.Tags, ","),
		"skip_checking":      strconv.FormatBool(options.SkipChecking),
		"paused":             strconv.FormatBool(options.Paused),
		"root_folder":        strconv.FormatBool(options.RootFolder),
		"rename":             options.Rename,
		"upLimit":            strconv.FormatInt(int64(options.UpLimit), 10),
		"dlLimit":            strconv.FormatInt(int64(options.DlLimit), 10),
		"ratioLimit":         strconv.FormatFloat(options.RatioLimit, 'f', 2, 32),
		"seedingTimeLimit":   strconv.FormatInt(int64(options.SeedingLimit), 10),
		"autoTMM":            strconv.FormatBool(options.AutoTMM),
		"sequentialDownload": strconv.FormatBool(options.SequentialDownload),
		"firstLastPiecePrio": strconv.FormatBool(options.FirstLastPiecePrio),
	}
}

const (
	FilterAllTorrentList                = "all"
	FilterDownloadingTorrentList        = "downloading"
	FilterSeedingTorrentList            = "seeding"
	FilterCompletedTorrentList          = "complete"
	FilterPausedTorrentList             = "paused"
	FilterActiveTorrentList             = "active"
	FilterInActiveTorrentList           = "inactive"
	FilterResumedTorrentList            = "resumed"
	FilterStalledTorrentList            = "stalled"
	FilterStalledUploadingTorrentList   = "stalled_uploading"
	FilterStalledDownloadingTorrentList = "stalled_downloading"
	FilterErroredTorrentList            = "errored"
)

const (
	StateError              = "error"
	StateMissingFile        = "missingFiles"
	StateUploading          = "uploading"
	StatePausedUP           = "pausedUP"
	StateQueueUP            = "queuedUP"
	StateStalledUP          = "stalledUP"
	StateCheckingUP         = "checkingUP"
	StateForcedUP           = "forcedUP"
	StateAllocating         = "allocating"
	StateDownloading        = "downloading"
	StateMetaDL             = "metaDL"
	StatePausedDL           = "pausedDL"
	StateQueuedDL           = "queuedDL"
	StateStalledDL          = "stalledDL"
	StateCheckingDL         = "checkingDL"
	StateCheckingResumeData = "checkingResumeData"
	StateMoving             = "moving"
	StateUnknown            = "unknown"
)

type RequestTorrentList struct {
	Filter   string
	Category string
	Tag      string
	Sort     string
	Reverse  bool
	Offset   int
	Limit    int
	Hashes   []string
}

func (req *RequestTorrentList) toMap() map[string]string {
	return map[string]string{
		"filter":   req.Filter,
		"category": req.Category,
		"tag":      req.Tag,
		"sort":     req.Sort,
		"reverse":  strconv.FormatBool(req.Reverse),
		"limit":    strconv.FormatInt(int64(req.Limit), 10),
		"offset":   strconv.FormatInt(int64(req.Offset), 10),
		"hashes":   strings.Join(req.Hashes, "|"),
	}
}

type Torrent struct {
	AddedOn           int     `json:"added_on"`
	AmountLeft        int     `json:"amount_left"`
	AutoTmm           bool    `json:"auto_tmm"`
	Availability      float64 `json:"availability"`
	Category          string  `json:"category"`
	Completed         int     `json:"completed"`
	CompletionOn      int     `json:"completion_on"`
	ContentPath       string  `json:"content_path"`
	DlLimit           int     `json:"dl_limit"`
	Dlspeed           int     `json:"dlspeed"`
	DownloadPath      string  `json:"download_path"`
	Downloaded        int     `json:"downloaded"`
	DownloadedSession int     `json:"downloaded_session"`
	Eta               int     `json:"eta"`
	FLPiecePrio       bool    `json:"f_l_piece_prio"`
	ForceStart        bool    `json:"force_start"`
	Hash              string  `json:"hash"`
	InfohashV1        string  `json:"infohash_v1"`
	InfohashV2        string  `json:"infohash_v2"`
	LastActivity      int     `json:"last_activity"`
	MagnetURI         string  `json:"magnet_uri"`
	MaxRatio          int     `json:"max_ratio"`
	MaxSeedingTime    int     `json:"max_seeding_time"`
	Name              string  `json:"name"`
	NumComplete       int     `json:"num_complete"`
	NumIncomplete     int     `json:"num_incomplete"`
	NumLeechs         int     `json:"num_leechs"`
	NumSeeds          int     `json:"num_seeds"`
	Priority          int     `json:"priority"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	RatioLimit        int     `json:"ratio_limit"`
	SavePath          string  `json:"save_path"`
	SeedingTime       int     `json:"seeding_time"`
	SeedingTimeLimit  int     `json:"seeding_time_limit"`
	SeenComplete      int     `json:"seen_complete"`
	SeqDl             bool    `json:"seq_dl"`
	Size              int     `json:"size"`
	State             string  `json:"state"`
	SuperSeeding      bool    `json:"super_seeding"`
	Tags              string  `json:"tags"`
	TimeActive        int     `json:"time_active"`
	TotalSize         int     `json:"total_size"`
	Tracker           string  `json:"tracker"`
	TrackersCount     int     `json:"trackers_count"`
	UpLimit           int     `json:"up_limit"`
	Uploaded          int     `json:"uploaded"`
	UploadedSession   int     `json:"uploaded_session"`
	Upspeed           int     `json:"upspeed"`
}

type TorrentProperties struct {
	AdditionDate           int     `json:"addition_date"`
	Comment                string  `json:"comment"`
	CompletionDate         int     `json:"completion_date"`
	CreatedBy              string  `json:"created_by"`
	CreationDate           int     `json:"creation_date"`
	DlLimit                int     `json:"dl_limit"`
	DlSpeed                int     `json:"dl_speed"`
	DlSpeedAvg             int     `json:"dl_speed_avg"`
	Eta                    int     `json:"eta"`
	LastSeen               int     `json:"last_seen"`
	NbConnections          int     `json:"nb_connections"`
	NbConnectionsLimit     int     `json:"nb_connections_limit"`
	Peers                  int     `json:"peers"`
	PeersTotal             int     `json:"peers_total"`
	PieceSize              int     `json:"piece_size"`
	PiecesHave             int     `json:"pieces_have"`
	PiecesNum              int     `json:"pieces_num"`
	Reannounce             int     `json:"reannounce"`
	SavePath               string  `json:"save_path"`
	SeedingTime            int     `json:"seeding_time"`
	Seeds                  int     `json:"seeds"`
	SeedsTotal             int     `json:"seeds_total"`
	ShareRatio             float64 `json:"share_ratio"`
	TimeElapsed            int     `json:"time_elapsed"`
	TotalDownloaded        int     `json:"total_downloaded"`
	TotalDownloadedSession int     `json:"total_downloaded_session"`
	TotalSize              int     `json:"total_size"`
	TotalUploaded          int     `json:"total_uploaded"`
	TotalUploadedSession   int     `json:"total_uploaded_session"`
	TotalWasted            int     `json:"total_wasted"`
	UpLimit                int     `json:"up_limit"`
	UpSpeed                int     `json:"up_speed"`
	UpSpeedAvg             int     `json:"up_speed_avg"`
}

type TorrentContent struct {
	Availability int     `json:"availability"`
	Index        int     `json:"index"`
	IsSeed       bool    `json:"is_seed"`
	Name         string  `json:"name"`
	PieceRange   []int   `json:"piece_range"`
	Priority     int     `json:"priority"`
	Progress     float64 `json:"progress"`
	Size         int     `json:"size"`
}

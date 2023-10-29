package jellyfin

import "time"

type ReqAuthenticateUserByName struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

type RespAuthenticateUserByName struct {
	User struct {
		Name                      string    `json:"Name"`
		ServerID                  string    `json:"ServerId"`
		ServerName                string    `json:"ServerName"`
		ID                        string    `json:"Id"`
		PrimaryImageTag           string    `json:"PrimaryImageTag"`
		HasPassword               bool      `json:"HasPassword"`
		HasConfiguredPassword     bool      `json:"HasConfiguredPassword"`
		HasConfiguredEasyPassword bool      `json:"HasConfiguredEasyPassword"`
		EnableAutoLogin           bool      `json:"EnableAutoLogin"`
		LastLoginDate             time.Time `json:"LastLoginDate"`
		LastActivityDate          time.Time `json:"LastActivityDate"`
		Configuration             struct {
			AudioLanguagePreference    string   `json:"AudioLanguagePreference"`
			PlayDefaultAudioTrack      bool     `json:"PlayDefaultAudioTrack"`
			SubtitleLanguagePreference string   `json:"SubtitleLanguagePreference"`
			DisplayMissingEpisodes     bool     `json:"DisplayMissingEpisodes"`
			GroupedFolders             []string `json:"GroupedFolders"`
			SubtitleMode               string   `json:"SubtitleMode"`
			DisplayCollectionsView     bool     `json:"DisplayCollectionsView"`
			EnableLocalPassword        bool     `json:"EnableLocalPassword"`
			OrderedViews               []string `json:"OrderedViews"`
			LatestItemsExcludes        []string `json:"LatestItemsExcludes"`
			MyMediaExcludes            []string `json:"MyMediaExcludes"`
			HidePlayedInLatest         bool     `json:"HidePlayedInLatest"`
			RememberAudioSelections    bool     `json:"RememberAudioSelections"`
			RememberSubtitleSelections bool     `json:"RememberSubtitleSelections"`
			EnableNextEpisodeAutoPlay  bool     `json:"EnableNextEpisodeAutoPlay"`
		} `json:"Configuration"`
		Policy struct {
			IsAdministrator            bool     `json:"IsAdministrator"`
			IsHidden                   bool     `json:"IsHidden"`
			IsDisabled                 bool     `json:"IsDisabled"`
			MaxParentalRating          int      `json:"MaxParentalRating"`
			BlockedTags                []string `json:"BlockedTags"`
			EnableUserPreferenceAccess bool     `json:"EnableUserPreferenceAccess"`
			AccessSchedules            []struct {
				ID        int    `json:"Id"`
				UserID    string `json:"UserId"`
				DayOfWeek string `json:"DayOfWeek"`
				StartHour int    `json:"StartHour"`
				EndHour   int    `json:"EndHour"`
			} `json:"AccessSchedules"`
			BlockUnratedItems                []string `json:"BlockUnratedItems"`
			EnableRemoteControlOfOtherUsers  bool     `json:"EnableRemoteControlOfOtherUsers"`
			EnableSharedDeviceControl        bool     `json:"EnableSharedDeviceControl"`
			EnableRemoteAccess               bool     `json:"EnableRemoteAccess"`
			EnableLiveTvManagement           bool     `json:"EnableLiveTvManagement"`
			EnableLiveTvAccess               bool     `json:"EnableLiveTvAccess"`
			EnableMediaPlayback              bool     `json:"EnableMediaPlayback"`
			EnableAudioPlaybackTranscoding   bool     `json:"EnableAudioPlaybackTranscoding"`
			EnableVideoPlaybackTranscoding   bool     `json:"EnableVideoPlaybackTranscoding"`
			EnablePlaybackRemuxing           bool     `json:"EnablePlaybackRemuxing"`
			ForceRemoteSourceTranscoding     bool     `json:"ForceRemoteSourceTranscoding"`
			EnableContentDeletion            bool     `json:"EnableContentDeletion"`
			EnableContentDeletionFromFolders []string `json:"EnableContentDeletionFromFolders"`
			EnableContentDownloading         bool     `json:"EnableContentDownloading"`
			EnableSyncTranscoding            bool     `json:"EnableSyncTranscoding"`
			EnableMediaConversion            bool     `json:"EnableMediaConversion"`
			EnabledDevices                   []string `json:"EnabledDevices"`
			EnableAllDevices                 bool     `json:"EnableAllDevices"`
			EnabledChannels                  []string `json:"EnabledChannels"`
			EnableAllChannels                bool     `json:"EnableAllChannels"`
			EnabledFolders                   []string `json:"EnabledFolders"`
			EnableAllFolders                 bool     `json:"EnableAllFolders"`
			InvalidLoginAttemptCount         int      `json:"InvalidLoginAttemptCount"`
			LoginAttemptsBeforeLockout       int      `json:"LoginAttemptsBeforeLockout"`
			MaxActiveSessions                int      `json:"MaxActiveSessions"`
			EnablePublicSharing              bool     `json:"EnablePublicSharing"`
			BlockedMediaFolders              []string `json:"BlockedMediaFolders"`
			BlockedChannels                  []string `json:"BlockedChannels"`
			RemoteClientBitrateLimit         int      `json:"RemoteClientBitrateLimit"`
			AuthenticationProviderID         string   `json:"AuthenticationProviderId"`
			PasswordResetProviderID          string   `json:"PasswordResetProviderId"`
			SyncPlayAccess                   string   `json:"SyncPlayAccess"`
		} `json:"Policy"`
		PrimaryImageAspectRatio int `json:"PrimaryImageAspectRatio"`
	} `json:"User"`
	SessionInfo struct {
		PlayState struct {
			PositionTicks       int    `json:"PositionTicks"`
			CanSeek             bool   `json:"CanSeek"`
			IsPaused            bool   `json:"IsPaused"`
			IsMuted             bool   `json:"IsMuted"`
			VolumeLevel         int    `json:"VolumeLevel"`
			AudioStreamIndex    int    `json:"AudioStreamIndex"`
			SubtitleStreamIndex int    `json:"SubtitleStreamIndex"`
			MediaSourceID       string `json:"MediaSourceId"`
			PlayMethod          string `json:"PlayMethod"`
			RepeatMode          string `json:"RepeatMode"`
			LiveStreamID        string `json:"LiveStreamId"`
		} `json:"PlayState"`
		AdditionalUsers []struct {
			UserID   string `json:"UserId"`
			UserName string `json:"UserName"`
		} `json:"AdditionalUsers"`
		Capabilities struct {
			PlayableMediaTypes           []string `json:"PlayableMediaTypes"`
			SupportedCommands            []string `json:"SupportedCommands"`
			SupportsMediaControl         bool     `json:"SupportsMediaControl"`
			SupportsContentUploading     bool     `json:"SupportsContentUploading"`
			MessageCallbackURL           string   `json:"MessageCallbackUrl"`
			SupportsPersistentIdentifier bool     `json:"SupportsPersistentIdentifier"`
			SupportsSync                 bool     `json:"SupportsSync"`
			DeviceProfile                struct {
				Name           string `json:"Name"`
				ID             string `json:"Id"`
				Identification struct {
					FriendlyName     string `json:"FriendlyName"`
					ModelNumber      string `json:"ModelNumber"`
					SerialNumber     string `json:"SerialNumber"`
					ModelName        string `json:"ModelName"`
					ModelDescription string `json:"ModelDescription"`
					ModelURL         string `json:"ModelUrl"`
					Manufacturer     string `json:"Manufacturer"`
					ManufacturerURL  string `json:"ManufacturerUrl"`
					Headers          []struct {
						Name  string `json:"Name"`
						Value string `json:"Value"`
						Match string `json:"Match"`
					} `json:"Headers"`
				} `json:"Identification"`
				FriendlyName                     string `json:"FriendlyName"`
				Manufacturer                     string `json:"Manufacturer"`
				ManufacturerURL                  string `json:"ManufacturerUrl"`
				ModelName                        string `json:"ModelName"`
				ModelDescription                 string `json:"ModelDescription"`
				ModelNumber                      string `json:"ModelNumber"`
				ModelURL                         string `json:"ModelUrl"`
				SerialNumber                     string `json:"SerialNumber"`
				EnableAlbumArtInDidl             bool   `json:"EnableAlbumArtInDidl"`
				EnableSingleAlbumArtLimit        bool   `json:"EnableSingleAlbumArtLimit"`
				EnableSingleSubtitleLimit        bool   `json:"EnableSingleSubtitleLimit"`
				SupportedMediaTypes              string `json:"SupportedMediaTypes"`
				UserID                           string `json:"UserId"`
				AlbumArtPn                       string `json:"AlbumArtPn"`
				MaxAlbumArtWidth                 int    `json:"MaxAlbumArtWidth"`
				MaxAlbumArtHeight                int    `json:"MaxAlbumArtHeight"`
				MaxIconWidth                     int    `json:"MaxIconWidth"`
				MaxIconHeight                    int    `json:"MaxIconHeight"`
				MaxStreamingBitrate              int    `json:"MaxStreamingBitrate"`
				MaxStaticBitrate                 int    `json:"MaxStaticBitrate"`
				MusicStreamingTranscodingBitrate int    `json:"MusicStreamingTranscodingBitrate"`
				MaxStaticMusicBitrate            int    `json:"MaxStaticMusicBitrate"`
				SonyAggregationFlags             string `json:"SonyAggregationFlags"`
				ProtocolInfo                     string `json:"ProtocolInfo"`
				TimelineOffsetSeconds            int    `json:"TimelineOffsetSeconds"`
				RequiresPlainVideoItems          bool   `json:"RequiresPlainVideoItems"`
				RequiresPlainFolders             bool   `json:"RequiresPlainFolders"`
				EnableMSMediaReceiverRegistrar   bool   `json:"EnableMSMediaReceiverRegistrar"`
				IgnoreTranscodeByteRangeRequests bool   `json:"IgnoreTranscodeByteRangeRequests"`
				XMLRootAttributes                []struct {
					Name  string `json:"Name"`
					Value string `json:"Value"`
				} `json:"XmlRootAttributes"`
				DirectPlayProfiles []struct {
					Container  string `json:"Container"`
					AudioCodec string `json:"AudioCodec"`
					VideoCodec string `json:"VideoCodec"`
					Type       string `json:"Type"`
				} `json:"DirectPlayProfiles"`
				TranscodingProfiles []struct {
					Container                 string `json:"Container"`
					Type                      string `json:"Type"`
					VideoCodec                string `json:"VideoCodec"`
					AudioCodec                string `json:"AudioCodec"`
					Protocol                  string `json:"Protocol"`
					EstimateContentLength     bool   `json:"EstimateContentLength"`
					EnableMpegtsM2TsMode      bool   `json:"EnableMpegtsM2TsMode"`
					TranscodeSeekInfo         string `json:"TranscodeSeekInfo"`
					CopyTimestamps            bool   `json:"CopyTimestamps"`
					Context                   string `json:"Context"`
					EnableSubtitlesInManifest bool   `json:"EnableSubtitlesInManifest"`
					MaxAudioChannels          string `json:"MaxAudioChannels"`
					MinSegments               int    `json:"MinSegments"`
					SegmentLength             int    `json:"SegmentLength"`
					BreakOnNonKeyFrames       bool   `json:"BreakOnNonKeyFrames"`
					Conditions                []struct {
						Condition  string `json:"Condition"`
						Property   string `json:"Property"`
						Value      string `json:"Value"`
						IsRequired bool   `json:"IsRequired"`
					} `json:"Conditions"`
				} `json:"TranscodingProfiles"`
				ContainerProfiles []struct {
					Type       string `json:"Type"`
					Conditions []struct {
						Condition  string `json:"Condition"`
						Property   string `json:"Property"`
						Value      string `json:"Value"`
						IsRequired bool   `json:"IsRequired"`
					} `json:"Conditions"`
					Container string `json:"Container"`
				} `json:"ContainerProfiles"`
				CodecProfiles []struct {
					Type       string `json:"Type"`
					Conditions []struct {
						Condition  string `json:"Condition"`
						Property   string `json:"Property"`
						Value      string `json:"Value"`
						IsRequired bool   `json:"IsRequired"`
					} `json:"Conditions"`
					ApplyConditions []struct {
						Condition  string `json:"Condition"`
						Property   string `json:"Property"`
						Value      string `json:"Value"`
						IsRequired bool   `json:"IsRequired"`
					} `json:"ApplyConditions"`
					Codec     string `json:"Codec"`
					Container string `json:"Container"`
				} `json:"CodecProfiles"`
				ResponseProfiles []struct {
					Container  string `json:"Container"`
					AudioCodec string `json:"AudioCodec"`
					VideoCodec string `json:"VideoCodec"`
					Type       string `json:"Type"`
					OrgPn      string `json:"OrgPn"`
					MimeType   string `json:"MimeType"`
					Conditions []struct {
						Condition  string `json:"Condition"`
						Property   string `json:"Property"`
						Value      string `json:"Value"`
						IsRequired bool   `json:"IsRequired"`
					} `json:"Conditions"`
				} `json:"ResponseProfiles"`
				SubtitleProfiles []struct {
					Format    string `json:"Format"`
					Method    string `json:"Method"`
					DidlMode  string `json:"DidlMode"`
					Language  string `json:"Language"`
					Container string `json:"Container"`
				} `json:"SubtitleProfiles"`
			} `json:"DeviceProfile"`
			AppStoreURL string `json:"AppStoreUrl"`
			IconURL     string `json:"IconUrl"`
		} `json:"Capabilities"`
		RemoteEndPoint      string    `json:"RemoteEndPoint"`
		PlayableMediaTypes  []string  `json:"PlayableMediaTypes"`
		ID                  string    `json:"Id"`
		UserID              string    `json:"UserId"`
		UserName            string    `json:"UserName"`
		Client              string    `json:"Client"`
		LastActivityDate    time.Time `json:"LastActivityDate"`
		LastPlaybackCheckIn time.Time `json:"LastPlaybackCheckIn"`
		DeviceName          string    `json:"DeviceName"`
		DeviceType          string    `json:"DeviceType"`
		NowPlayingItem      struct {
			Name                         string    `json:"Name"`
			OriginalTitle                string    `json:"OriginalTitle"`
			ServerID                     string    `json:"ServerId"`
			ID                           string    `json:"Id"`
			Etag                         string    `json:"Etag"`
			SourceType                   string    `json:"SourceType"`
			PlaylistItemID               string    `json:"PlaylistItemId"`
			DateCreated                  time.Time `json:"DateCreated"`
			DateLastMediaAdded           time.Time `json:"DateLastMediaAdded"`
			ExtraType                    string    `json:"ExtraType"`
			AirsBeforeSeasonNumber       int       `json:"AirsBeforeSeasonNumber"`
			AirsAfterSeasonNumber        int       `json:"AirsAfterSeasonNumber"`
			AirsBeforeEpisodeNumber      int       `json:"AirsBeforeEpisodeNumber"`
			CanDelete                    bool      `json:"CanDelete"`
			CanDownload                  bool      `json:"CanDownload"`
			HasSubtitles                 bool      `json:"HasSubtitles"`
			PreferredMetadataLanguage    string    `json:"PreferredMetadataLanguage"`
			PreferredMetadataCountryCode string    `json:"PreferredMetadataCountryCode"`
			SupportsSync                 bool      `json:"SupportsSync"`
			Container                    string    `json:"Container"`
			SortName                     string    `json:"SortName"`
			ForcedSortName               string    `json:"ForcedSortName"`
			Video3DFormat                string    `json:"Video3DFormat"`
			PremiereDate                 time.Time `json:"PremiereDate"`
			ExternalUrls                 []struct {
				Name string `json:"Name"`
				URL  string `json:"Url"`
			} `json:"ExternalUrls"`
			MediaSources []struct {
				Protocol              string `json:"Protocol"`
				ID                    string `json:"Id"`
				Path                  string `json:"Path"`
				EncoderPath           string `json:"EncoderPath"`
				EncoderProtocol       string `json:"EncoderProtocol"`
				Type                  string `json:"Type"`
				Container             string `json:"Container"`
				Size                  int    `json:"Size"`
				Name                  string `json:"Name"`
				IsRemote              bool   `json:"IsRemote"`
				ETag                  string `json:"ETag"`
				RunTimeTicks          int    `json:"RunTimeTicks"`
				ReadAtNativeFramerate bool   `json:"ReadAtNativeFramerate"`
				IgnoreDts             bool   `json:"IgnoreDts"`
				IgnoreIndex           bool   `json:"IgnoreIndex"`
				GenPtsInput           bool   `json:"GenPtsInput"`
				SupportsTranscoding   bool   `json:"SupportsTranscoding"`
				SupportsDirectStream  bool   `json:"SupportsDirectStream"`
				SupportsDirectPlay    bool   `json:"SupportsDirectPlay"`
				IsInfiniteStream      bool   `json:"IsInfiniteStream"`
				RequiresOpening       bool   `json:"RequiresOpening"`
				OpenToken             string `json:"OpenToken"`
				RequiresClosing       bool   `json:"RequiresClosing"`
				LiveStreamID          string `json:"LiveStreamId"`
				BufferMs              int    `json:"BufferMs"`
				RequiresLooping       bool   `json:"RequiresLooping"`
				SupportsProbing       bool   `json:"SupportsProbing"`
				VideoType             string `json:"VideoType"`
				IsoType               string `json:"IsoType"`
				Video3DFormat         string `json:"Video3DFormat"`
				MediaStreams          []struct {
					Codec                     string `json:"Codec"`
					CodecTag                  string `json:"CodecTag"`
					Language                  string `json:"Language"`
					ColorRange                string `json:"ColorRange"`
					ColorSpace                string `json:"ColorSpace"`
					ColorTransfer             string `json:"ColorTransfer"`
					ColorPrimaries            string `json:"ColorPrimaries"`
					DvVersionMajor            int    `json:"DvVersionMajor"`
					DvVersionMinor            int    `json:"DvVersionMinor"`
					DvProfile                 int    `json:"DvProfile"`
					DvLevel                   int    `json:"DvLevel"`
					RpuPresentFlag            int    `json:"RpuPresentFlag"`
					ElPresentFlag             int    `json:"ElPresentFlag"`
					BlPresentFlag             int    `json:"BlPresentFlag"`
					DvBlSignalCompatibilityID int    `json:"DvBlSignalCompatibilityId"`
					Comment                   string `json:"Comment"`
					TimeBase                  string `json:"TimeBase"`
					CodecTimeBase             string `json:"CodecTimeBase"`
					Title                     string `json:"Title"`
					VideoRange                string `json:"VideoRange"`
					VideoRangeType            string `json:"VideoRangeType"`
					VideoDoViTitle            string `json:"VideoDoViTitle"`
					LocalizedUndefined        string `json:"LocalizedUndefined"`
					LocalizedDefault          string `json:"LocalizedDefault"`
					LocalizedForced           string `json:"LocalizedForced"`
					LocalizedExternal         string `json:"LocalizedExternal"`
					DisplayTitle              string `json:"DisplayTitle"`
					NalLengthSize             string `json:"NalLengthSize"`
					IsInterlaced              bool   `json:"IsInterlaced"`
					IsAVC                     bool   `json:"IsAVC"`
					ChannelLayout             string `json:"ChannelLayout"`
					BitRate                   int    `json:"BitRate"`
					BitDepth                  int    `json:"BitDepth"`
					RefFrames                 int    `json:"RefFrames"`
					PacketLength              int    `json:"PacketLength"`
					Channels                  int    `json:"Channels"`
					SampleRate                int    `json:"SampleRate"`
					IsDefault                 bool   `json:"IsDefault"`
					IsForced                  bool   `json:"IsForced"`
					Height                    int    `json:"Height"`
					Width                     int    `json:"Width"`
					AverageFrameRate          int    `json:"AverageFrameRate"`
					RealFrameRate             int    `json:"RealFrameRate"`
					Profile                   string `json:"Profile"`
					Type                      string `json:"Type"`
					AspectRatio               string `json:"AspectRatio"`
					Index                     int    `json:"Index"`
					Score                     int    `json:"Score"`
					IsExternal                bool   `json:"IsExternal"`
					DeliveryMethod            string `json:"DeliveryMethod"`
					DeliveryURL               string `json:"DeliveryUrl"`
					IsExternalURL             bool   `json:"IsExternalUrl"`
					IsTextSubtitleStream      bool   `json:"IsTextSubtitleStream"`
					SupportsExternalStream    bool   `json:"SupportsExternalStream"`
					Path                      string `json:"Path"`
					PixelFormat               string `json:"PixelFormat"`
					Level                     int    `json:"Level"`
					IsAnamorphic              bool   `json:"IsAnamorphic"`
				} `json:"MediaStreams"`
				MediaAttachments []struct {
					Codec       string `json:"Codec"`
					CodecTag    string `json:"CodecTag"`
					Comment     string `json:"Comment"`
					Index       int    `json:"Index"`
					FileName    string `json:"FileName"`
					MimeType    string `json:"MimeType"`
					DeliveryURL string `json:"DeliveryUrl"`
				} `json:"MediaAttachments"`
				Formats             []string `json:"Formats"`
				Bitrate             int      `json:"Bitrate"`
				Timestamp           string   `json:"Timestamp"`
				RequiredHTTPHeaders struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"RequiredHttpHeaders"`
				TranscodingURL             string `json:"TranscodingUrl"`
				TranscodingSubProtocol     string `json:"TranscodingSubProtocol"`
				TranscodingContainer       string `json:"TranscodingContainer"`
				AnalyzeDurationMs          int    `json:"AnalyzeDurationMs"`
				DefaultAudioStreamIndex    int    `json:"DefaultAudioStreamIndex"`
				DefaultSubtitleStreamIndex int    `json:"DefaultSubtitleStreamIndex"`
			} `json:"MediaSources"`
			CriticRating             int      `json:"CriticRating"`
			ProductionLocations      []string `json:"ProductionLocations"`
			Path                     string   `json:"Path"`
			EnableMediaSourceDisplay bool     `json:"EnableMediaSourceDisplay"`
			OfficialRating           string   `json:"OfficialRating"`
			CustomRating             string   `json:"CustomRating"`
			ChannelID                string   `json:"ChannelId"`
			ChannelName              string   `json:"ChannelName"`
			Overview                 string   `json:"Overview"`
			Taglines                 []string `json:"Taglines"`
			Genres                   []string `json:"Genres"`
			CommunityRating          int      `json:"CommunityRating"`
			CumulativeRunTimeTicks   int      `json:"CumulativeRunTimeTicks"`
			RunTimeTicks             int      `json:"RunTimeTicks"`
			PlayAccess               string   `json:"PlayAccess"`
			AspectRatio              string   `json:"AspectRatio"`
			ProductionYear           int      `json:"ProductionYear"`
			IsPlaceHolder            bool     `json:"IsPlaceHolder"`
			Number                   string   `json:"Number"`
			ChannelNumber            string   `json:"ChannelNumber"`
			IndexNumber              int      `json:"IndexNumber"`
			IndexNumberEnd           int      `json:"IndexNumberEnd"`
			ParentIndexNumber        int      `json:"ParentIndexNumber"`
			RemoteTrailers           []struct {
				URL  string `json:"Url"`
				Name string `json:"Name"`
			} `json:"RemoteTrailers"`
			ProviderIds struct {
				Property1 string `json:"property1"`
				Property2 string `json:"property2"`
			} `json:"ProviderIds"`
			IsHD     bool   `json:"IsHD"`
			IsFolder bool   `json:"IsFolder"`
			ParentID string `json:"ParentId"`
			Type     string `json:"Type"`
			People   []struct {
				Name            string `json:"Name"`
				ID              string `json:"Id"`
				Role            string `json:"Role"`
				Type            string `json:"Type"`
				PrimaryImageTag string `json:"PrimaryImageTag"`
				ImageBlurHashes struct {
					Primary struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Primary"`
					Art struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Art"`
					Backdrop struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Backdrop"`
					Banner struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Banner"`
					Logo struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Logo"`
					Thumb struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Thumb"`
					Disc struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Disc"`
					Box struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Box"`
					Screenshot struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Screenshot"`
					Menu struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Menu"`
					Chapter struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Chapter"`
					BoxRear struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"BoxRear"`
					Profile struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Profile"`
				} `json:"ImageBlurHashes"`
			} `json:"People"`
			Studios []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"Studios"`
			GenreItems []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"GenreItems"`
			ParentLogoItemID        string   `json:"ParentLogoItemId"`
			ParentBackdropItemID    string   `json:"ParentBackdropItemId"`
			ParentBackdropImageTags []string `json:"ParentBackdropImageTags"`
			LocalTrailerCount       int      `json:"LocalTrailerCount"`
			UserData                struct {
				Rating                int       `json:"Rating"`
				PlayedPercentage      int       `json:"PlayedPercentage"`
				UnplayedItemCount     int       `json:"UnplayedItemCount"`
				PlaybackPositionTicks int       `json:"PlaybackPositionTicks"`
				PlayCount             int       `json:"PlayCount"`
				IsFavorite            bool      `json:"IsFavorite"`
				Likes                 bool      `json:"Likes"`
				LastPlayedDate        time.Time `json:"LastPlayedDate"`
				Played                bool      `json:"Played"`
				Key                   string    `json:"Key"`
				ItemID                string    `json:"ItemId"`
			} `json:"UserData"`
			RecursiveItemCount      int      `json:"RecursiveItemCount"`
			ChildCount              int      `json:"ChildCount"`
			SeriesName              string   `json:"SeriesName"`
			SeriesID                string   `json:"SeriesId"`
			SeasonID                string   `json:"SeasonId"`
			SpecialFeatureCount     int      `json:"SpecialFeatureCount"`
			DisplayPreferencesID    string   `json:"DisplayPreferencesId"`
			Status                  string   `json:"Status"`
			AirTime                 string   `json:"AirTime"`
			AirDays                 []string `json:"AirDays"`
			Tags                    []string `json:"Tags"`
			PrimaryImageAspectRatio int      `json:"PrimaryImageAspectRatio"`
			Artists                 []string `json:"Artists"`
			ArtistItems             []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"ArtistItems"`
			Album                 string `json:"Album"`
			CollectionType        string `json:"CollectionType"`
			DisplayOrder          string `json:"DisplayOrder"`
			AlbumID               string `json:"AlbumId"`
			AlbumPrimaryImageTag  string `json:"AlbumPrimaryImageTag"`
			SeriesPrimaryImageTag string `json:"SeriesPrimaryImageTag"`
			AlbumArtist           string `json:"AlbumArtist"`
			AlbumArtists          []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"AlbumArtists"`
			SeasonName   string `json:"SeasonName"`
			MediaStreams []struct {
				Codec                     string `json:"Codec"`
				CodecTag                  string `json:"CodecTag"`
				Language                  string `json:"Language"`
				ColorRange                string `json:"ColorRange"`
				ColorSpace                string `json:"ColorSpace"`
				ColorTransfer             string `json:"ColorTransfer"`
				ColorPrimaries            string `json:"ColorPrimaries"`
				DvVersionMajor            int    `json:"DvVersionMajor"`
				DvVersionMinor            int    `json:"DvVersionMinor"`
				DvProfile                 int    `json:"DvProfile"`
				DvLevel                   int    `json:"DvLevel"`
				RpuPresentFlag            int    `json:"RpuPresentFlag"`
				ElPresentFlag             int    `json:"ElPresentFlag"`
				BlPresentFlag             int    `json:"BlPresentFlag"`
				DvBlSignalCompatibilityID int    `json:"DvBlSignalCompatibilityId"`
				Comment                   string `json:"Comment"`
				TimeBase                  string `json:"TimeBase"`
				CodecTimeBase             string `json:"CodecTimeBase"`
				Title                     string `json:"Title"`
				VideoRange                string `json:"VideoRange"`
				VideoRangeType            string `json:"VideoRangeType"`
				VideoDoViTitle            string `json:"VideoDoViTitle"`
				LocalizedUndefined        string `json:"LocalizedUndefined"`
				LocalizedDefault          string `json:"LocalizedDefault"`
				LocalizedForced           string `json:"LocalizedForced"`
				LocalizedExternal         string `json:"LocalizedExternal"`
				DisplayTitle              string `json:"DisplayTitle"`
				NalLengthSize             string `json:"NalLengthSize"`
				IsInterlaced              bool   `json:"IsInterlaced"`
				IsAVC                     bool   `json:"IsAVC"`
				ChannelLayout             string `json:"ChannelLayout"`
				BitRate                   int    `json:"BitRate"`
				BitDepth                  int    `json:"BitDepth"`
				RefFrames                 int    `json:"RefFrames"`
				PacketLength              int    `json:"PacketLength"`
				Channels                  int    `json:"Channels"`
				SampleRate                int    `json:"SampleRate"`
				IsDefault                 bool   `json:"IsDefault"`
				IsForced                  bool   `json:"IsForced"`
				Height                    int    `json:"Height"`
				Width                     int    `json:"Width"`
				AverageFrameRate          int    `json:"AverageFrameRate"`
				RealFrameRate             int    `json:"RealFrameRate"`
				Profile                   string `json:"Profile"`
				Type                      string `json:"Type"`
				AspectRatio               string `json:"AspectRatio"`
				Index                     int    `json:"Index"`
				Score                     int    `json:"Score"`
				IsExternal                bool   `json:"IsExternal"`
				DeliveryMethod            string `json:"DeliveryMethod"`
				DeliveryURL               string `json:"DeliveryUrl"`
				IsExternalURL             bool   `json:"IsExternalUrl"`
				IsTextSubtitleStream      bool   `json:"IsTextSubtitleStream"`
				SupportsExternalStream    bool   `json:"SupportsExternalStream"`
				Path                      string `json:"Path"`
				PixelFormat               string `json:"PixelFormat"`
				Level                     int    `json:"Level"`
				IsAnamorphic              bool   `json:"IsAnamorphic"`
			} `json:"MediaStreams"`
			VideoType        string `json:"VideoType"`
			PartCount        int    `json:"PartCount"`
			MediaSourceCount int    `json:"MediaSourceCount"`
			ImageTags        struct {
				Property1 string `json:"property1"`
				Property2 string `json:"property2"`
			} `json:"ImageTags"`
			BackdropImageTags   []string `json:"BackdropImageTags"`
			ScreenshotImageTags []string `json:"ScreenshotImageTags"`
			ParentLogoImageTag  string   `json:"ParentLogoImageTag"`
			ParentArtItemID     string   `json:"ParentArtItemId"`
			ParentArtImageTag   string   `json:"ParentArtImageTag"`
			SeriesThumbImageTag string   `json:"SeriesThumbImageTag"`
			ImageBlurHashes     struct {
				Primary struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Primary"`
				Art struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Art"`
				Backdrop struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Backdrop"`
				Banner struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Banner"`
				Logo struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Logo"`
				Thumb struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Thumb"`
				Disc struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Disc"`
				Box struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Box"`
				Screenshot struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Screenshot"`
				Menu struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Menu"`
				Chapter struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Chapter"`
				BoxRear struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"BoxRear"`
				Profile struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Profile"`
			} `json:"ImageBlurHashes"`
			SeriesStudio             string `json:"SeriesStudio"`
			ParentThumbItemID        string `json:"ParentThumbItemId"`
			ParentThumbImageTag      string `json:"ParentThumbImageTag"`
			ParentPrimaryImageItemID string `json:"ParentPrimaryImageItemId"`
			ParentPrimaryImageTag    string `json:"ParentPrimaryImageTag"`
			Chapters                 []struct {
				StartPositionTicks int       `json:"StartPositionTicks"`
				Name               string    `json:"Name"`
				ImagePath          string    `json:"ImagePath"`
				ImageDateModified  time.Time `json:"ImageDateModified"`
				ImageTag           string    `json:"ImageTag"`
			} `json:"Chapters"`
			LocationType           string    `json:"LocationType"`
			IsoType                string    `json:"IsoType"`
			MediaType              string    `json:"MediaType"`
			EndDate                time.Time `json:"EndDate"`
			LockedFields           []string  `json:"LockedFields"`
			TrailerCount           int       `json:"TrailerCount"`
			MovieCount             int       `json:"MovieCount"`
			SeriesCount            int       `json:"SeriesCount"`
			ProgramCount           int       `json:"ProgramCount"`
			EpisodeCount           int       `json:"EpisodeCount"`
			SongCount              int       `json:"SongCount"`
			AlbumCount             int       `json:"AlbumCount"`
			ArtistCount            int       `json:"ArtistCount"`
			MusicVideoCount        int       `json:"MusicVideoCount"`
			LockData               bool      `json:"LockData"`
			Width                  int       `json:"Width"`
			Height                 int       `json:"Height"`
			CameraMake             string    `json:"CameraMake"`
			CameraModel            string    `json:"CameraModel"`
			Software               string    `json:"Software"`
			ExposureTime           int       `json:"ExposureTime"`
			FocalLength            int       `json:"FocalLength"`
			ImageOrientation       string    `json:"ImageOrientation"`
			Aperture               int       `json:"Aperture"`
			ShutterSpeed           int       `json:"ShutterSpeed"`
			Latitude               int       `json:"Latitude"`
			Longitude              int       `json:"Longitude"`
			Altitude               int       `json:"Altitude"`
			IsoSpeedRating         int       `json:"IsoSpeedRating"`
			SeriesTimerID          string    `json:"SeriesTimerId"`
			ProgramID              string    `json:"ProgramId"`
			ChannelPrimaryImageTag string    `json:"ChannelPrimaryImageTag"`
			StartDate              time.Time `json:"StartDate"`
			CompletionPercentage   int       `json:"CompletionPercentage"`
			IsRepeat               bool      `json:"IsRepeat"`
			EpisodeTitle           string    `json:"EpisodeTitle"`
			ChannelType            string    `json:"ChannelType"`
			Audio                  string    `json:"Audio"`
			IsMovie                bool      `json:"IsMovie"`
			IsSports               bool      `json:"IsSports"`
			IsSeries               bool      `json:"IsSeries"`
			IsLive                 bool      `json:"IsLive"`
			IsNews                 bool      `json:"IsNews"`
			IsKids                 bool      `json:"IsKids"`
			IsPremiere             bool      `json:"IsPremiere"`
			TimerID                string    `json:"TimerId"`
			CurrentProgram         struct {
			} `json:"CurrentProgram"`
		} `json:"NowPlayingItem"`
		FullNowPlayingItem struct {
			Size           int       `json:"Size"`
			Container      string    `json:"Container"`
			IsHD           bool      `json:"IsHD"`
			IsShortcut     bool      `json:"IsShortcut"`
			ShortcutPath   string    `json:"ShortcutPath"`
			Width          int       `json:"Width"`
			Height         int       `json:"Height"`
			ExtraIds       []string  `json:"ExtraIds"`
			DateLastSaved  time.Time `json:"DateLastSaved"`
			RemoteTrailers []struct {
				URL  string `json:"Url"`
				Name string `json:"Name"`
			} `json:"RemoteTrailers"`
			SupportsExternalTransfer bool `json:"SupportsExternalTransfer"`
		} `json:"FullNowPlayingItem"`
		NowViewingItem struct {
			Name                         string    `json:"Name"`
			OriginalTitle                string    `json:"OriginalTitle"`
			ServerID                     string    `json:"ServerId"`
			ID                           string    `json:"Id"`
			Etag                         string    `json:"Etag"`
			SourceType                   string    `json:"SourceType"`
			PlaylistItemID               string    `json:"PlaylistItemId"`
			DateCreated                  time.Time `json:"DateCreated"`
			DateLastMediaAdded           time.Time `json:"DateLastMediaAdded"`
			ExtraType                    string    `json:"ExtraType"`
			AirsBeforeSeasonNumber       int       `json:"AirsBeforeSeasonNumber"`
			AirsAfterSeasonNumber        int       `json:"AirsAfterSeasonNumber"`
			AirsBeforeEpisodeNumber      int       `json:"AirsBeforeEpisodeNumber"`
			CanDelete                    bool      `json:"CanDelete"`
			CanDownload                  bool      `json:"CanDownload"`
			HasSubtitles                 bool      `json:"HasSubtitles"`
			PreferredMetadataLanguage    string    `json:"PreferredMetadataLanguage"`
			PreferredMetadataCountryCode string    `json:"PreferredMetadataCountryCode"`
			SupportsSync                 bool      `json:"SupportsSync"`
			Container                    string    `json:"Container"`
			SortName                     string    `json:"SortName"`
			ForcedSortName               string    `json:"ForcedSortName"`
			Video3DFormat                string    `json:"Video3DFormat"`
			PremiereDate                 time.Time `json:"PremiereDate"`
			ExternalUrls                 []struct {
				Name string `json:"Name"`
				URL  string `json:"Url"`
			} `json:"ExternalUrls"`
			MediaSources []struct {
				Protocol              string `json:"Protocol"`
				ID                    string `json:"Id"`
				Path                  string `json:"Path"`
				EncoderPath           string `json:"EncoderPath"`
				EncoderProtocol       string `json:"EncoderProtocol"`
				Type                  string `json:"Type"`
				Container             string `json:"Container"`
				Size                  int    `json:"Size"`
				Name                  string `json:"Name"`
				IsRemote              bool   `json:"IsRemote"`
				ETag                  string `json:"ETag"`
				RunTimeTicks          int    `json:"RunTimeTicks"`
				ReadAtNativeFramerate bool   `json:"ReadAtNativeFramerate"`
				IgnoreDts             bool   `json:"IgnoreDts"`
				IgnoreIndex           bool   `json:"IgnoreIndex"`
				GenPtsInput           bool   `json:"GenPtsInput"`
				SupportsTranscoding   bool   `json:"SupportsTranscoding"`
				SupportsDirectStream  bool   `json:"SupportsDirectStream"`
				SupportsDirectPlay    bool   `json:"SupportsDirectPlay"`
				IsInfiniteStream      bool   `json:"IsInfiniteStream"`
				RequiresOpening       bool   `json:"RequiresOpening"`
				OpenToken             string `json:"OpenToken"`
				RequiresClosing       bool   `json:"RequiresClosing"`
				LiveStreamID          string `json:"LiveStreamId"`
				BufferMs              int    `json:"BufferMs"`
				RequiresLooping       bool   `json:"RequiresLooping"`
				SupportsProbing       bool   `json:"SupportsProbing"`
				VideoType             string `json:"VideoType"`
				IsoType               string `json:"IsoType"`
				Video3DFormat         string `json:"Video3DFormat"`
				MediaStreams          []struct {
					Codec                     string `json:"Codec"`
					CodecTag                  string `json:"CodecTag"`
					Language                  string `json:"Language"`
					ColorRange                string `json:"ColorRange"`
					ColorSpace                string `json:"ColorSpace"`
					ColorTransfer             string `json:"ColorTransfer"`
					ColorPrimaries            string `json:"ColorPrimaries"`
					DvVersionMajor            int    `json:"DvVersionMajor"`
					DvVersionMinor            int    `json:"DvVersionMinor"`
					DvProfile                 int    `json:"DvProfile"`
					DvLevel                   int    `json:"DvLevel"`
					RpuPresentFlag            int    `json:"RpuPresentFlag"`
					ElPresentFlag             int    `json:"ElPresentFlag"`
					BlPresentFlag             int    `json:"BlPresentFlag"`
					DvBlSignalCompatibilityID int    `json:"DvBlSignalCompatibilityId"`
					Comment                   string `json:"Comment"`
					TimeBase                  string `json:"TimeBase"`
					CodecTimeBase             string `json:"CodecTimeBase"`
					Title                     string `json:"Title"`
					VideoRange                string `json:"VideoRange"`
					VideoRangeType            string `json:"VideoRangeType"`
					VideoDoViTitle            string `json:"VideoDoViTitle"`
					LocalizedUndefined        string `json:"LocalizedUndefined"`
					LocalizedDefault          string `json:"LocalizedDefault"`
					LocalizedForced           string `json:"LocalizedForced"`
					LocalizedExternal         string `json:"LocalizedExternal"`
					DisplayTitle              string `json:"DisplayTitle"`
					NalLengthSize             string `json:"NalLengthSize"`
					IsInterlaced              bool   `json:"IsInterlaced"`
					IsAVC                     bool   `json:"IsAVC"`
					ChannelLayout             string `json:"ChannelLayout"`
					BitRate                   int    `json:"BitRate"`
					BitDepth                  int    `json:"BitDepth"`
					RefFrames                 int    `json:"RefFrames"`
					PacketLength              int    `json:"PacketLength"`
					Channels                  int    `json:"Channels"`
					SampleRate                int    `json:"SampleRate"`
					IsDefault                 bool   `json:"IsDefault"`
					IsForced                  bool   `json:"IsForced"`
					Height                    int    `json:"Height"`
					Width                     int    `json:"Width"`
					AverageFrameRate          int    `json:"AverageFrameRate"`
					RealFrameRate             int    `json:"RealFrameRate"`
					Profile                   string `json:"Profile"`
					Type                      string `json:"Type"`
					AspectRatio               string `json:"AspectRatio"`
					Index                     int    `json:"Index"`
					Score                     int    `json:"Score"`
					IsExternal                bool   `json:"IsExternal"`
					DeliveryMethod            string `json:"DeliveryMethod"`
					DeliveryURL               string `json:"DeliveryUrl"`
					IsExternalURL             bool   `json:"IsExternalUrl"`
					IsTextSubtitleStream      bool   `json:"IsTextSubtitleStream"`
					SupportsExternalStream    bool   `json:"SupportsExternalStream"`
					Path                      string `json:"Path"`
					PixelFormat               string `json:"PixelFormat"`
					Level                     int    `json:"Level"`
					IsAnamorphic              bool   `json:"IsAnamorphic"`
				} `json:"MediaStreams"`
				MediaAttachments []struct {
					Codec       string `json:"Codec"`
					CodecTag    string `json:"CodecTag"`
					Comment     string `json:"Comment"`
					Index       int    `json:"Index"`
					FileName    string `json:"FileName"`
					MimeType    string `json:"MimeType"`
					DeliveryURL string `json:"DeliveryUrl"`
				} `json:"MediaAttachments"`
				Formats             []string `json:"Formats"`
				Bitrate             int      `json:"Bitrate"`
				Timestamp           string   `json:"Timestamp"`
				RequiredHTTPHeaders struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"RequiredHttpHeaders"`
				TranscodingURL             string `json:"TranscodingUrl"`
				TranscodingSubProtocol     string `json:"TranscodingSubProtocol"`
				TranscodingContainer       string `json:"TranscodingContainer"`
				AnalyzeDurationMs          int    `json:"AnalyzeDurationMs"`
				DefaultAudioStreamIndex    int    `json:"DefaultAudioStreamIndex"`
				DefaultSubtitleStreamIndex int    `json:"DefaultSubtitleStreamIndex"`
			} `json:"MediaSources"`
			CriticRating             int      `json:"CriticRating"`
			ProductionLocations      []string `json:"ProductionLocations"`
			Path                     string   `json:"Path"`
			EnableMediaSourceDisplay bool     `json:"EnableMediaSourceDisplay"`
			OfficialRating           string   `json:"OfficialRating"`
			CustomRating             string   `json:"CustomRating"`
			ChannelID                string   `json:"ChannelId"`
			ChannelName              string   `json:"ChannelName"`
			Overview                 string   `json:"Overview"`
			Taglines                 []string `json:"Taglines"`
			Genres                   []string `json:"Genres"`
			CommunityRating          int      `json:"CommunityRating"`
			CumulativeRunTimeTicks   int      `json:"CumulativeRunTimeTicks"`
			RunTimeTicks             int      `json:"RunTimeTicks"`
			PlayAccess               string   `json:"PlayAccess"`
			AspectRatio              string   `json:"AspectRatio"`
			ProductionYear           int      `json:"ProductionYear"`
			IsPlaceHolder            bool     `json:"IsPlaceHolder"`
			Number                   string   `json:"Number"`
			ChannelNumber            string   `json:"ChannelNumber"`
			IndexNumber              int      `json:"IndexNumber"`
			IndexNumberEnd           int      `json:"IndexNumberEnd"`
			ParentIndexNumber        int      `json:"ParentIndexNumber"`
			RemoteTrailers           []struct {
				URL  string `json:"Url"`
				Name string `json:"Name"`
			} `json:"RemoteTrailers"`
			ProviderIds struct {
				Property1 string `json:"property1"`
				Property2 string `json:"property2"`
			} `json:"ProviderIds"`
			IsHD     bool   `json:"IsHD"`
			IsFolder bool   `json:"IsFolder"`
			ParentID string `json:"ParentId"`
			Type     string `json:"Type"`
			People   []struct {
				Name            string `json:"Name"`
				ID              string `json:"Id"`
				Role            string `json:"Role"`
				Type            string `json:"Type"`
				PrimaryImageTag string `json:"PrimaryImageTag"`
				ImageBlurHashes struct {
					Primary struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Primary"`
					Art struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Art"`
					Backdrop struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Backdrop"`
					Banner struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Banner"`
					Logo struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Logo"`
					Thumb struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Thumb"`
					Disc struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Disc"`
					Box struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Box"`
					Screenshot struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Screenshot"`
					Menu struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Menu"`
					Chapter struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Chapter"`
					BoxRear struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"BoxRear"`
					Profile struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Profile"`
				} `json:"ImageBlurHashes"`
			} `json:"People"`
			Studios []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"Studios"`
			GenreItems []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"GenreItems"`
			ParentLogoItemID        string   `json:"ParentLogoItemId"`
			ParentBackdropItemID    string   `json:"ParentBackdropItemId"`
			ParentBackdropImageTags []string `json:"ParentBackdropImageTags"`
			LocalTrailerCount       int      `json:"LocalTrailerCount"`
			UserData                struct {
				Rating                int       `json:"Rating"`
				PlayedPercentage      int       `json:"PlayedPercentage"`
				UnplayedItemCount     int       `json:"UnplayedItemCount"`
				PlaybackPositionTicks int       `json:"PlaybackPositionTicks"`
				PlayCount             int       `json:"PlayCount"`
				IsFavorite            bool      `json:"IsFavorite"`
				Likes                 bool      `json:"Likes"`
				LastPlayedDate        time.Time `json:"LastPlayedDate"`
				Played                bool      `json:"Played"`
				Key                   string    `json:"Key"`
				ItemID                string    `json:"ItemId"`
			} `json:"UserData"`
			RecursiveItemCount      int      `json:"RecursiveItemCount"`
			ChildCount              int      `json:"ChildCount"`
			SeriesName              string   `json:"SeriesName"`
			SeriesID                string   `json:"SeriesId"`
			SeasonID                string   `json:"SeasonId"`
			SpecialFeatureCount     int      `json:"SpecialFeatureCount"`
			DisplayPreferencesID    string   `json:"DisplayPreferencesId"`
			Status                  string   `json:"Status"`
			AirTime                 string   `json:"AirTime"`
			AirDays                 []string `json:"AirDays"`
			Tags                    []string `json:"Tags"`
			PrimaryImageAspectRatio int      `json:"PrimaryImageAspectRatio"`
			Artists                 []string `json:"Artists"`
			ArtistItems             []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"ArtistItems"`
			Album                 string `json:"Album"`
			CollectionType        string `json:"CollectionType"`
			DisplayOrder          string `json:"DisplayOrder"`
			AlbumID               string `json:"AlbumId"`
			AlbumPrimaryImageTag  string `json:"AlbumPrimaryImageTag"`
			SeriesPrimaryImageTag string `json:"SeriesPrimaryImageTag"`
			AlbumArtist           string `json:"AlbumArtist"`
			AlbumArtists          []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"AlbumArtists"`
			SeasonName   string `json:"SeasonName"`
			MediaStreams []struct {
				Codec                     string `json:"Codec"`
				CodecTag                  string `json:"CodecTag"`
				Language                  string `json:"Language"`
				ColorRange                string `json:"ColorRange"`
				ColorSpace                string `json:"ColorSpace"`
				ColorTransfer             string `json:"ColorTransfer"`
				ColorPrimaries            string `json:"ColorPrimaries"`
				DvVersionMajor            int    `json:"DvVersionMajor"`
				DvVersionMinor            int    `json:"DvVersionMinor"`
				DvProfile                 int    `json:"DvProfile"`
				DvLevel                   int    `json:"DvLevel"`
				RpuPresentFlag            int    `json:"RpuPresentFlag"`
				ElPresentFlag             int    `json:"ElPresentFlag"`
				BlPresentFlag             int    `json:"BlPresentFlag"`
				DvBlSignalCompatibilityID int    `json:"DvBlSignalCompatibilityId"`
				Comment                   string `json:"Comment"`
				TimeBase                  string `json:"TimeBase"`
				CodecTimeBase             string `json:"CodecTimeBase"`
				Title                     string `json:"Title"`
				VideoRange                string `json:"VideoRange"`
				VideoRangeType            string `json:"VideoRangeType"`
				VideoDoViTitle            string `json:"VideoDoViTitle"`
				LocalizedUndefined        string `json:"LocalizedUndefined"`
				LocalizedDefault          string `json:"LocalizedDefault"`
				LocalizedForced           string `json:"LocalizedForced"`
				LocalizedExternal         string `json:"LocalizedExternal"`
				DisplayTitle              string `json:"DisplayTitle"`
				NalLengthSize             string `json:"NalLengthSize"`
				IsInterlaced              bool   `json:"IsInterlaced"`
				IsAVC                     bool   `json:"IsAVC"`
				ChannelLayout             string `json:"ChannelLayout"`
				BitRate                   int    `json:"BitRate"`
				BitDepth                  int    `json:"BitDepth"`
				RefFrames                 int    `json:"RefFrames"`
				PacketLength              int    `json:"PacketLength"`
				Channels                  int    `json:"Channels"`
				SampleRate                int    `json:"SampleRate"`
				IsDefault                 bool   `json:"IsDefault"`
				IsForced                  bool   `json:"IsForced"`
				Height                    int    `json:"Height"`
				Width                     int    `json:"Width"`
				AverageFrameRate          int    `json:"AverageFrameRate"`
				RealFrameRate             int    `json:"RealFrameRate"`
				Profile                   string `json:"Profile"`
				Type                      string `json:"Type"`
				AspectRatio               string `json:"AspectRatio"`
				Index                     int    `json:"Index"`
				Score                     int    `json:"Score"`
				IsExternal                bool   `json:"IsExternal"`
				DeliveryMethod            string `json:"DeliveryMethod"`
				DeliveryURL               string `json:"DeliveryUrl"`
				IsExternalURL             bool   `json:"IsExternalUrl"`
				IsTextSubtitleStream      bool   `json:"IsTextSubtitleStream"`
				SupportsExternalStream    bool   `json:"SupportsExternalStream"`
				Path                      string `json:"Path"`
				PixelFormat               string `json:"PixelFormat"`
				Level                     int    `json:"Level"`
				IsAnamorphic              bool   `json:"IsAnamorphic"`
			} `json:"MediaStreams"`
			VideoType        string `json:"VideoType"`
			PartCount        int    `json:"PartCount"`
			MediaSourceCount int    `json:"MediaSourceCount"`
			ImageTags        struct {
				Property1 string `json:"property1"`
				Property2 string `json:"property2"`
			} `json:"ImageTags"`
			BackdropImageTags   []string `json:"BackdropImageTags"`
			ScreenshotImageTags []string `json:"ScreenshotImageTags"`
			ParentLogoImageTag  string   `json:"ParentLogoImageTag"`
			ParentArtItemID     string   `json:"ParentArtItemId"`
			ParentArtImageTag   string   `json:"ParentArtImageTag"`
			SeriesThumbImageTag string   `json:"SeriesThumbImageTag"`
			ImageBlurHashes     struct {
				Primary struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Primary"`
				Art struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Art"`
				Backdrop struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Backdrop"`
				Banner struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Banner"`
				Logo struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Logo"`
				Thumb struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Thumb"`
				Disc struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Disc"`
				Box struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Box"`
				Screenshot struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Screenshot"`
				Menu struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Menu"`
				Chapter struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Chapter"`
				BoxRear struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"BoxRear"`
				Profile struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Profile"`
			} `json:"ImageBlurHashes"`
			SeriesStudio             string `json:"SeriesStudio"`
			ParentThumbItemID        string `json:"ParentThumbItemId"`
			ParentThumbImageTag      string `json:"ParentThumbImageTag"`
			ParentPrimaryImageItemID string `json:"ParentPrimaryImageItemId"`
			ParentPrimaryImageTag    string `json:"ParentPrimaryImageTag"`
			Chapters                 []struct {
				StartPositionTicks int       `json:"StartPositionTicks"`
				Name               string    `json:"Name"`
				ImagePath          string    `json:"ImagePath"`
				ImageDateModified  time.Time `json:"ImageDateModified"`
				ImageTag           string    `json:"ImageTag"`
			} `json:"Chapters"`
			LocationType           string    `json:"LocationType"`
			IsoType                string    `json:"IsoType"`
			MediaType              string    `json:"MediaType"`
			EndDate                time.Time `json:"EndDate"`
			LockedFields           []string  `json:"LockedFields"`
			TrailerCount           int       `json:"TrailerCount"`
			MovieCount             int       `json:"MovieCount"`
			SeriesCount            int       `json:"SeriesCount"`
			ProgramCount           int       `json:"ProgramCount"`
			EpisodeCount           int       `json:"EpisodeCount"`
			SongCount              int       `json:"SongCount"`
			AlbumCount             int       `json:"AlbumCount"`
			ArtistCount            int       `json:"ArtistCount"`
			MusicVideoCount        int       `json:"MusicVideoCount"`
			LockData               bool      `json:"LockData"`
			Width                  int       `json:"Width"`
			Height                 int       `json:"Height"`
			CameraMake             string    `json:"CameraMake"`
			CameraModel            string    `json:"CameraModel"`
			Software               string    `json:"Software"`
			ExposureTime           int       `json:"ExposureTime"`
			FocalLength            int       `json:"FocalLength"`
			ImageOrientation       string    `json:"ImageOrientation"`
			Aperture               int       `json:"Aperture"`
			ShutterSpeed           int       `json:"ShutterSpeed"`
			Latitude               int       `json:"Latitude"`
			Longitude              int       `json:"Longitude"`
			Altitude               int       `json:"Altitude"`
			IsoSpeedRating         int       `json:"IsoSpeedRating"`
			SeriesTimerID          string    `json:"SeriesTimerId"`
			ProgramID              string    `json:"ProgramId"`
			ChannelPrimaryImageTag string    `json:"ChannelPrimaryImageTag"`
			StartDate              time.Time `json:"StartDate"`
			CompletionPercentage   int       `json:"CompletionPercentage"`
			IsRepeat               bool      `json:"IsRepeat"`
			EpisodeTitle           string    `json:"EpisodeTitle"`
			ChannelType            string    `json:"ChannelType"`
			Audio                  string    `json:"Audio"`
			IsMovie                bool      `json:"IsMovie"`
			IsSports               bool      `json:"IsSports"`
			IsSeries               bool      `json:"IsSeries"`
			IsLive                 bool      `json:"IsLive"`
			IsNews                 bool      `json:"IsNews"`
			IsKids                 bool      `json:"IsKids"`
			IsPremiere             bool      `json:"IsPremiere"`
			TimerID                string    `json:"TimerId"`
			CurrentProgram         struct {
			} `json:"CurrentProgram"`
		} `json:"NowViewingItem"`
		DeviceID           string `json:"DeviceId"`
		ApplicationVersion string `json:"ApplicationVersion"`
		TranscodingInfo    struct {
			AudioCodec               string   `json:"AudioCodec"`
			VideoCodec               string   `json:"VideoCodec"`
			Container                string   `json:"Container"`
			IsVideoDirect            bool     `json:"IsVideoDirect"`
			IsAudioDirect            bool     `json:"IsAudioDirect"`
			Bitrate                  int      `json:"Bitrate"`
			Framerate                int      `json:"Framerate"`
			CompletionPercentage     int      `json:"CompletionPercentage"`
			Width                    int      `json:"Width"`
			Height                   int      `json:"Height"`
			AudioChannels            int      `json:"AudioChannels"`
			HardwareAccelerationType string   `json:"HardwareAccelerationType"`
			TranscodeReasons         []string `json:"TranscodeReasons"`
		} `json:"TranscodingInfo"`
		IsActive              bool `json:"IsActive"`
		SupportsMediaControl  bool `json:"SupportsMediaControl"`
		SupportsRemoteControl bool `json:"SupportsRemoteControl"`
		NowPlayingQueue       []struct {
			ID             string `json:"Id"`
			PlaylistItemID string `json:"PlaylistItemId"`
		} `json:"NowPlayingQueue"`
		NowPlayingQueueFullItems []struct {
			Name                         string    `json:"Name"`
			OriginalTitle                string    `json:"OriginalTitle"`
			ServerID                     string    `json:"ServerId"`
			ID                           string    `json:"Id"`
			Etag                         string    `json:"Etag"`
			SourceType                   string    `json:"SourceType"`
			PlaylistItemID               string    `json:"PlaylistItemId"`
			DateCreated                  time.Time `json:"DateCreated"`
			DateLastMediaAdded           time.Time `json:"DateLastMediaAdded"`
			ExtraType                    string    `json:"ExtraType"`
			AirsBeforeSeasonNumber       int       `json:"AirsBeforeSeasonNumber"`
			AirsAfterSeasonNumber        int       `json:"AirsAfterSeasonNumber"`
			AirsBeforeEpisodeNumber      int       `json:"AirsBeforeEpisodeNumber"`
			CanDelete                    bool      `json:"CanDelete"`
			CanDownload                  bool      `json:"CanDownload"`
			HasSubtitles                 bool      `json:"HasSubtitles"`
			PreferredMetadataLanguage    string    `json:"PreferredMetadataLanguage"`
			PreferredMetadataCountryCode string    `json:"PreferredMetadataCountryCode"`
			SupportsSync                 bool      `json:"SupportsSync"`
			Container                    string    `json:"Container"`
			SortName                     string    `json:"SortName"`
			ForcedSortName               string    `json:"ForcedSortName"`
			Video3DFormat                string    `json:"Video3DFormat"`
			PremiereDate                 time.Time `json:"PremiereDate"`
			ExternalUrls                 []struct {
				Name string `json:"Name"`
				URL  string `json:"Url"`
			} `json:"ExternalUrls"`
			MediaSources []struct {
				Protocol              string `json:"Protocol"`
				ID                    string `json:"Id"`
				Path                  string `json:"Path"`
				EncoderPath           string `json:"EncoderPath"`
				EncoderProtocol       string `json:"EncoderProtocol"`
				Type                  string `json:"Type"`
				Container             string `json:"Container"`
				Size                  int    `json:"Size"`
				Name                  string `json:"Name"`
				IsRemote              bool   `json:"IsRemote"`
				ETag                  string `json:"ETag"`
				RunTimeTicks          int    `json:"RunTimeTicks"`
				ReadAtNativeFramerate bool   `json:"ReadAtNativeFramerate"`
				IgnoreDts             bool   `json:"IgnoreDts"`
				IgnoreIndex           bool   `json:"IgnoreIndex"`
				GenPtsInput           bool   `json:"GenPtsInput"`
				SupportsTranscoding   bool   `json:"SupportsTranscoding"`
				SupportsDirectStream  bool   `json:"SupportsDirectStream"`
				SupportsDirectPlay    bool   `json:"SupportsDirectPlay"`
				IsInfiniteStream      bool   `json:"IsInfiniteStream"`
				RequiresOpening       bool   `json:"RequiresOpening"`
				OpenToken             string `json:"OpenToken"`
				RequiresClosing       bool   `json:"RequiresClosing"`
				LiveStreamID          string `json:"LiveStreamId"`
				BufferMs              int    `json:"BufferMs"`
				RequiresLooping       bool   `json:"RequiresLooping"`
				SupportsProbing       bool   `json:"SupportsProbing"`
				VideoType             string `json:"VideoType"`
				IsoType               string `json:"IsoType"`
				Video3DFormat         string `json:"Video3DFormat"`
				MediaStreams          []struct {
					Codec                     string `json:"Codec"`
					CodecTag                  string `json:"CodecTag"`
					Language                  string `json:"Language"`
					ColorRange                string `json:"ColorRange"`
					ColorSpace                string `json:"ColorSpace"`
					ColorTransfer             string `json:"ColorTransfer"`
					ColorPrimaries            string `json:"ColorPrimaries"`
					DvVersionMajor            int    `json:"DvVersionMajor"`
					DvVersionMinor            int    `json:"DvVersionMinor"`
					DvProfile                 int    `json:"DvProfile"`
					DvLevel                   int    `json:"DvLevel"`
					RpuPresentFlag            int    `json:"RpuPresentFlag"`
					ElPresentFlag             int    `json:"ElPresentFlag"`
					BlPresentFlag             int    `json:"BlPresentFlag"`
					DvBlSignalCompatibilityID int    `json:"DvBlSignalCompatibilityId"`
					Comment                   string `json:"Comment"`
					TimeBase                  string `json:"TimeBase"`
					CodecTimeBase             string `json:"CodecTimeBase"`
					Title                     string `json:"Title"`
					VideoRange                string `json:"VideoRange"`
					VideoRangeType            string `json:"VideoRangeType"`
					VideoDoViTitle            string `json:"VideoDoViTitle"`
					LocalizedUndefined        string `json:"LocalizedUndefined"`
					LocalizedDefault          string `json:"LocalizedDefault"`
					LocalizedForced           string `json:"LocalizedForced"`
					LocalizedExternal         string `json:"LocalizedExternal"`
					DisplayTitle              string `json:"DisplayTitle"`
					NalLengthSize             string `json:"NalLengthSize"`
					IsInterlaced              bool   `json:"IsInterlaced"`
					IsAVC                     bool   `json:"IsAVC"`
					ChannelLayout             string `json:"ChannelLayout"`
					BitRate                   int    `json:"BitRate"`
					BitDepth                  int    `json:"BitDepth"`
					RefFrames                 int    `json:"RefFrames"`
					PacketLength              int    `json:"PacketLength"`
					Channels                  int    `json:"Channels"`
					SampleRate                int    `json:"SampleRate"`
					IsDefault                 bool   `json:"IsDefault"`
					IsForced                  bool   `json:"IsForced"`
					Height                    int    `json:"Height"`
					Width                     int    `json:"Width"`
					AverageFrameRate          int    `json:"AverageFrameRate"`
					RealFrameRate             int    `json:"RealFrameRate"`
					Profile                   string `json:"Profile"`
					Type                      string `json:"Type"`
					AspectRatio               string `json:"AspectRatio"`
					Index                     int    `json:"Index"`
					Score                     int    `json:"Score"`
					IsExternal                bool   `json:"IsExternal"`
					DeliveryMethod            string `json:"DeliveryMethod"`
					DeliveryURL               string `json:"DeliveryUrl"`
					IsExternalURL             bool   `json:"IsExternalUrl"`
					IsTextSubtitleStream      bool   `json:"IsTextSubtitleStream"`
					SupportsExternalStream    bool   `json:"SupportsExternalStream"`
					Path                      string `json:"Path"`
					PixelFormat               string `json:"PixelFormat"`
					Level                     int    `json:"Level"`
					IsAnamorphic              bool   `json:"IsAnamorphic"`
				} `json:"MediaStreams"`
				MediaAttachments []struct {
					Codec       string `json:"Codec"`
					CodecTag    string `json:"CodecTag"`
					Comment     string `json:"Comment"`
					Index       int    `json:"Index"`
					FileName    string `json:"FileName"`
					MimeType    string `json:"MimeType"`
					DeliveryURL string `json:"DeliveryUrl"`
				} `json:"MediaAttachments"`
				Formats             []string `json:"Formats"`
				Bitrate             int      `json:"Bitrate"`
				Timestamp           string   `json:"Timestamp"`
				RequiredHTTPHeaders struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"RequiredHttpHeaders"`
				TranscodingURL             string `json:"TranscodingUrl"`
				TranscodingSubProtocol     string `json:"TranscodingSubProtocol"`
				TranscodingContainer       string `json:"TranscodingContainer"`
				AnalyzeDurationMs          int    `json:"AnalyzeDurationMs"`
				DefaultAudioStreamIndex    int    `json:"DefaultAudioStreamIndex"`
				DefaultSubtitleStreamIndex int    `json:"DefaultSubtitleStreamIndex"`
			} `json:"MediaSources"`
			CriticRating             int      `json:"CriticRating"`
			ProductionLocations      []string `json:"ProductionLocations"`
			Path                     string   `json:"Path"`
			EnableMediaSourceDisplay bool     `json:"EnableMediaSourceDisplay"`
			OfficialRating           string   `json:"OfficialRating"`
			CustomRating             string   `json:"CustomRating"`
			ChannelID                string   `json:"ChannelId"`
			ChannelName              string   `json:"ChannelName"`
			Overview                 string   `json:"Overview"`
			Taglines                 []string `json:"Taglines"`
			Genres                   []string `json:"Genres"`
			CommunityRating          int      `json:"CommunityRating"`
			CumulativeRunTimeTicks   int      `json:"CumulativeRunTimeTicks"`
			RunTimeTicks             int      `json:"RunTimeTicks"`
			PlayAccess               string   `json:"PlayAccess"`
			AspectRatio              string   `json:"AspectRatio"`
			ProductionYear           int      `json:"ProductionYear"`
			IsPlaceHolder            bool     `json:"IsPlaceHolder"`
			Number                   string   `json:"Number"`
			ChannelNumber            string   `json:"ChannelNumber"`
			IndexNumber              int      `json:"IndexNumber"`
			IndexNumberEnd           int      `json:"IndexNumberEnd"`
			ParentIndexNumber        int      `json:"ParentIndexNumber"`
			RemoteTrailers           []struct {
				URL  string `json:"Url"`
				Name string `json:"Name"`
			} `json:"RemoteTrailers"`
			ProviderIds struct {
				Property1 string `json:"property1"`
				Property2 string `json:"property2"`
			} `json:"ProviderIds"`
			IsHD     bool   `json:"IsHD"`
			IsFolder bool   `json:"IsFolder"`
			ParentID string `json:"ParentId"`
			Type     string `json:"Type"`
			People   []struct {
				Name            string `json:"Name"`
				ID              string `json:"Id"`
				Role            string `json:"Role"`
				Type            string `json:"Type"`
				PrimaryImageTag string `json:"PrimaryImageTag"`
				ImageBlurHashes struct {
					Primary struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Primary"`
					Art struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Art"`
					Backdrop struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Backdrop"`
					Banner struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Banner"`
					Logo struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Logo"`
					Thumb struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Thumb"`
					Disc struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Disc"`
					Box struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Box"`
					Screenshot struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Screenshot"`
					Menu struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Menu"`
					Chapter struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Chapter"`
					BoxRear struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"BoxRear"`
					Profile struct {
						Property1 string `json:"property1"`
						Property2 string `json:"property2"`
					} `json:"Profile"`
				} `json:"ImageBlurHashes"`
			} `json:"People"`
			Studios []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"Studios"`
			GenreItems []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"GenreItems"`
			ParentLogoItemID        string   `json:"ParentLogoItemId"`
			ParentBackdropItemID    string   `json:"ParentBackdropItemId"`
			ParentBackdropImageTags []string `json:"ParentBackdropImageTags"`
			LocalTrailerCount       int      `json:"LocalTrailerCount"`
			UserData                struct {
				Rating                int       `json:"Rating"`
				PlayedPercentage      int       `json:"PlayedPercentage"`
				UnplayedItemCount     int       `json:"UnplayedItemCount"`
				PlaybackPositionTicks int       `json:"PlaybackPositionTicks"`
				PlayCount             int       `json:"PlayCount"`
				IsFavorite            bool      `json:"IsFavorite"`
				Likes                 bool      `json:"Likes"`
				LastPlayedDate        time.Time `json:"LastPlayedDate"`
				Played                bool      `json:"Played"`
				Key                   string    `json:"Key"`
				ItemID                string    `json:"ItemId"`
			} `json:"UserData"`
			RecursiveItemCount      int      `json:"RecursiveItemCount"`
			ChildCount              int      `json:"ChildCount"`
			SeriesName              string   `json:"SeriesName"`
			SeriesID                string   `json:"SeriesId"`
			SeasonID                string   `json:"SeasonId"`
			SpecialFeatureCount     int      `json:"SpecialFeatureCount"`
			DisplayPreferencesID    string   `json:"DisplayPreferencesId"`
			Status                  string   `json:"Status"`
			AirTime                 string   `json:"AirTime"`
			AirDays                 []string `json:"AirDays"`
			Tags                    []string `json:"Tags"`
			PrimaryImageAspectRatio int      `json:"PrimaryImageAspectRatio"`
			Artists                 []string `json:"Artists"`
			ArtistItems             []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"ArtistItems"`
			Album                 string `json:"Album"`
			CollectionType        string `json:"CollectionType"`
			DisplayOrder          string `json:"DisplayOrder"`
			AlbumID               string `json:"AlbumId"`
			AlbumPrimaryImageTag  string `json:"AlbumPrimaryImageTag"`
			SeriesPrimaryImageTag string `json:"SeriesPrimaryImageTag"`
			AlbumArtist           string `json:"AlbumArtist"`
			AlbumArtists          []struct {
				Name string `json:"Name"`
				ID   string `json:"Id"`
			} `json:"AlbumArtists"`
			SeasonName   string `json:"SeasonName"`
			MediaStreams []struct {
				Codec                     string `json:"Codec"`
				CodecTag                  string `json:"CodecTag"`
				Language                  string `json:"Language"`
				ColorRange                string `json:"ColorRange"`
				ColorSpace                string `json:"ColorSpace"`
				ColorTransfer             string `json:"ColorTransfer"`
				ColorPrimaries            string `json:"ColorPrimaries"`
				DvVersionMajor            int    `json:"DvVersionMajor"`
				DvVersionMinor            int    `json:"DvVersionMinor"`
				DvProfile                 int    `json:"DvProfile"`
				DvLevel                   int    `json:"DvLevel"`
				RpuPresentFlag            int    `json:"RpuPresentFlag"`
				ElPresentFlag             int    `json:"ElPresentFlag"`
				BlPresentFlag             int    `json:"BlPresentFlag"`
				DvBlSignalCompatibilityID int    `json:"DvBlSignalCompatibilityId"`
				Comment                   string `json:"Comment"`
				TimeBase                  string `json:"TimeBase"`
				CodecTimeBase             string `json:"CodecTimeBase"`
				Title                     string `json:"Title"`
				VideoRange                string `json:"VideoRange"`
				VideoRangeType            string `json:"VideoRangeType"`
				VideoDoViTitle            string `json:"VideoDoViTitle"`
				LocalizedUndefined        string `json:"LocalizedUndefined"`
				LocalizedDefault          string `json:"LocalizedDefault"`
				LocalizedForced           string `json:"LocalizedForced"`
				LocalizedExternal         string `json:"LocalizedExternal"`
				DisplayTitle              string `json:"DisplayTitle"`
				NalLengthSize             string `json:"NalLengthSize"`
				IsInterlaced              bool   `json:"IsInterlaced"`
				IsAVC                     bool   `json:"IsAVC"`
				ChannelLayout             string `json:"ChannelLayout"`
				BitRate                   int    `json:"BitRate"`
				BitDepth                  int    `json:"BitDepth"`
				RefFrames                 int    `json:"RefFrames"`
				PacketLength              int    `json:"PacketLength"`
				Channels                  int    `json:"Channels"`
				SampleRate                int    `json:"SampleRate"`
				IsDefault                 bool   `json:"IsDefault"`
				IsForced                  bool   `json:"IsForced"`
				Height                    int    `json:"Height"`
				Width                     int    `json:"Width"`
				AverageFrameRate          int    `json:"AverageFrameRate"`
				RealFrameRate             int    `json:"RealFrameRate"`
				Profile                   string `json:"Profile"`
				Type                      string `json:"Type"`
				AspectRatio               string `json:"AspectRatio"`
				Index                     int    `json:"Index"`
				Score                     int    `json:"Score"`
				IsExternal                bool   `json:"IsExternal"`
				DeliveryMethod            string `json:"DeliveryMethod"`
				DeliveryURL               string `json:"DeliveryUrl"`
				IsExternalURL             bool   `json:"IsExternalUrl"`
				IsTextSubtitleStream      bool   `json:"IsTextSubtitleStream"`
				SupportsExternalStream    bool   `json:"SupportsExternalStream"`
				Path                      string `json:"Path"`
				PixelFormat               string `json:"PixelFormat"`
				Level                     int    `json:"Level"`
				IsAnamorphic              bool   `json:"IsAnamorphic"`
			} `json:"MediaStreams"`
			VideoType        string `json:"VideoType"`
			PartCount        int    `json:"PartCount"`
			MediaSourceCount int    `json:"MediaSourceCount"`
			ImageTags        struct {
				Property1 string `json:"property1"`
				Property2 string `json:"property2"`
			} `json:"ImageTags"`
			BackdropImageTags   []string `json:"BackdropImageTags"`
			ScreenshotImageTags []string `json:"ScreenshotImageTags"`
			ParentLogoImageTag  string   `json:"ParentLogoImageTag"`
			ParentArtItemID     string   `json:"ParentArtItemId"`
			ParentArtImageTag   string   `json:"ParentArtImageTag"`
			SeriesThumbImageTag string   `json:"SeriesThumbImageTag"`
			ImageBlurHashes     struct {
				Primary struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Primary"`
				Art struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Art"`
				Backdrop struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Backdrop"`
				Banner struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Banner"`
				Logo struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Logo"`
				Thumb struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Thumb"`
				Disc struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Disc"`
				Box struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Box"`
				Screenshot struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Screenshot"`
				Menu struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Menu"`
				Chapter struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Chapter"`
				BoxRear struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"BoxRear"`
				Profile struct {
					Property1 string `json:"property1"`
					Property2 string `json:"property2"`
				} `json:"Profile"`
			} `json:"ImageBlurHashes"`
			SeriesStudio             string `json:"SeriesStudio"`
			ParentThumbItemID        string `json:"ParentThumbItemId"`
			ParentThumbImageTag      string `json:"ParentThumbImageTag"`
			ParentPrimaryImageItemID string `json:"ParentPrimaryImageItemId"`
			ParentPrimaryImageTag    string `json:"ParentPrimaryImageTag"`
			Chapters                 []struct {
				StartPositionTicks int       `json:"StartPositionTicks"`
				Name               string    `json:"Name"`
				ImagePath          string    `json:"ImagePath"`
				ImageDateModified  time.Time `json:"ImageDateModified"`
				ImageTag           string    `json:"ImageTag"`
			} `json:"Chapters"`
			LocationType           string    `json:"LocationType"`
			IsoType                string    `json:"IsoType"`
			MediaType              string    `json:"MediaType"`
			EndDate                time.Time `json:"EndDate"`
			LockedFields           []string  `json:"LockedFields"`
			TrailerCount           int       `json:"TrailerCount"`
			MovieCount             int       `json:"MovieCount"`
			SeriesCount            int       `json:"SeriesCount"`
			ProgramCount           int       `json:"ProgramCount"`
			EpisodeCount           int       `json:"EpisodeCount"`
			SongCount              int       `json:"SongCount"`
			AlbumCount             int       `json:"AlbumCount"`
			ArtistCount            int       `json:"ArtistCount"`
			MusicVideoCount        int       `json:"MusicVideoCount"`
			LockData               bool      `json:"LockData"`
			Width                  int       `json:"Width"`
			Height                 int       `json:"Height"`
			CameraMake             string    `json:"CameraMake"`
			CameraModel            string    `json:"CameraModel"`
			Software               string    `json:"Software"`
			ExposureTime           int       `json:"ExposureTime"`
			FocalLength            int       `json:"FocalLength"`
			ImageOrientation       string    `json:"ImageOrientation"`
			Aperture               int       `json:"Aperture"`
			ShutterSpeed           int       `json:"ShutterSpeed"`
			Latitude               int       `json:"Latitude"`
			Longitude              int       `json:"Longitude"`
			Altitude               int       `json:"Altitude"`
			IsoSpeedRating         int       `json:"IsoSpeedRating"`
			SeriesTimerID          string    `json:"SeriesTimerId"`
			ProgramID              string    `json:"ProgramId"`
			ChannelPrimaryImageTag string    `json:"ChannelPrimaryImageTag"`
			StartDate              time.Time `json:"StartDate"`
			CompletionPercentage   int       `json:"CompletionPercentage"`
			IsRepeat               bool      `json:"IsRepeat"`
			EpisodeTitle           string    `json:"EpisodeTitle"`
			ChannelType            string    `json:"ChannelType"`
			Audio                  string    `json:"Audio"`
			IsMovie                bool      `json:"IsMovie"`
			IsSports               bool      `json:"IsSports"`
			IsSeries               bool      `json:"IsSeries"`
			IsLive                 bool      `json:"IsLive"`
			IsNews                 bool      `json:"IsNews"`
			IsKids                 bool      `json:"IsKids"`
			IsPremiere             bool      `json:"IsPremiere"`
			TimerID                string    `json:"TimerId"`
			CurrentProgram         struct {
			} `json:"CurrentProgram"`
		} `json:"NowPlayingQueueFullItems"`
		HasCustomDeviceName bool     `json:"HasCustomDeviceName"`
		PlaylistItemID      string   `json:"PlaylistItemId"`
		ServerID            string   `json:"ServerId"`
		UserPrimaryImageTag string   `json:"UserPrimaryImageTag"`
		SupportedCommands   []string `json:"SupportedCommands"`
	} `json:"SessionInfo"`
	AccessToken string `json:"AccessToken"`
	ServerID    string `json:"ServerId"`
}

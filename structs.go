package main

import "encoding/xml"

type Transport struct{}

type WriteCounter struct {
	Total      int64
	TotalStr   string
	Downloaded int64
	Percentage int
	StartTime  int64
}

type Config struct {
	Email         string
	Password      string
	Urls          []string
	Format        int
	FormatStr     string
	OutPath       string
	TrackTemplate string
	Lyrics        bool
}

type Args struct {
	Urls    []string `arg:"positional, required"`
	Format  int      `arg:"-f" default:"-1" help:"Download quality. 1 = AAC 128, 2 = AAC 320."`
	OutPath string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
	Lyrics  bool     `arg:"-l" help:"Get lyrics if available."`
}

type TrackMeta struct {
	TrackID        int    `json:"trackId"`
	MusicTitle     string `json:"musicTitle"`
	ArtistID       int    `json:"artistId"`
	ArtistName     string `json:"artistName"`
	AlbumID        int    `json:"albumId"`
	AlbumTitle     string `json:"albumTitle"`
	Tieup          string `json:"tieup"`
	ImageURL       string `json:"imageUrl"`
	DiscNo         int    `json:"discNo"`
	TrackNo        int    `json:"trackNo"`
	OnetimeURL     string `json:"onetimeUrl"`
	IsNapster      bool   `json:"isNapster"`
	IsFavorite     bool   `json:"isFavorite"`
	PlayTimeNormal int    `json:"playTimeNormal"`
	PlayTimeHigh   int    `json:"playTimeHigh"`
	DataSizeNormal int    `json:"dataSizeNormal"`
	DataSizeHigh   int    `json:"dataSizeHigh"`
}

type AlbumMeta struct {
	Result                int    `json:"result"`
	AlbumID               int    `json:"albumId"`
	AlbumTitle            string `json:"albumTitle"`
	ArtistID              int    `json:"artistId"`
	ArtistName            string `json:"artistName"`
	ImageURL              string `json:"imageUrl"`
	TrackCount            int    `json:"trackCount"`
	PlayTimeNormal        int    `json:"playTimeNormal"`
	PlayTimeHigh          int    `json:"playTimeHigh"`
	DataSizeNormal        int    `json:"dataSizeNormal"`
	DataSizeHigh          int    `json:"dataSizeHigh"`
	SalesDate             string `json:"salesDate"`
	IsFavorite            bool   `json:"isFavorite"`
	TowerRecordsOnlineURL string `json:"towerRecordsOnlineUrl"`
	DiscList              []struct {
		DiscNo    int         `json:"discNo"`
		Count     int         `json:"count"`
		TrackList []TrackMeta `json:"trackList"`
	} `json:"discList"`
	RelatedPlaylist struct {
		Count        int `json:"count"`
		NextID       int `json:"nextId"`
		PlaylistList []struct {
			PlaylistID   int    `json:"playlistId"`
			PlaylistName string `json:"playlistName"`
			Description  string `json:"description"`
			ImageURL     string `json:"imageUrl"`
			IsNewArrival bool   `json:"isNewArrival"`
		} `json:"playlistList"`
	} `json:"relatedPlaylist"`
	ArtistList []struct {
		ArtistID   int    `json:"artistId"`
		ArtistName string `json:"artistName"`
		ImageURL   string `json:"imageUrl"`
		Profile    string `json:"profile"`
	} `json:"artistList"`
}

type LyricsMeta struct {
	XMLName xml.Name `xml:"result"`
	Text    string   `xml:",chardata"`
	Head    struct {
		Text              string `xml:",chardata"`
		StatusCode        string `xml:"status-code"`
		Message           string `xml:"message"`
		RequestURI        string `xml:"request-uri"`
		RequestMethod     string `xml:"request-method"`
		RequestParameters struct {
			Text        string `xml:",chardata"`
			UserID      string `xml:"user-id"`
			Type        string `xml:"type"`
			AccessToken string `xml:"access-token"`
		} `xml:"request-parameters"`
		RequestDatetime string `xml:"request-datetime"`
		RequestToken    string `xml:"request-token"`
	} `xml:"head"`
	Body struct {
		Text string `xml:",chardata"`
		ID   string `xml:"id"`
		Type string `xml:"type"`
		Song struct {
			Text      string `xml:",chardata"`
			MgTrackID string `xml:"mg-track-id"`
			MscID     string `xml:"msc-id"`
			Title     string `xml:"title"`
			Artist    string `xml:"artist"`
			Lyricist  string `xml:"lyricist"`
			Composer  string `xml:"composer"`
			Time      string `xml:"time"`
		} `xml:"song"`
		Data string `xml:"data"`
	} `xml:"body"`
}

type Lyrics []struct {
	Time  string `json:"time"`
	Words string `json:"words"`
}

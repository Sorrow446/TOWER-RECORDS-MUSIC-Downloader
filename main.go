package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	ap "github.com/Sorrow446/go-atomicparsley"
	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
)

const (
	regexString = `^https://music.tower.jp/album/detail/(\d+)$`
	urlBase     = "https://music.tower.jp/"
	apiBase     = "https://api.music.tower.jp/api/v1/"
	lyricsUrl   = "https://yanone.music-sc.jp/lyrics/"
	accessToken = "e2b5b1656d97543321f5931992eebf08"
	userAgent   = "TRMUSIC ASUS_Z01QD(SP;9.9.9.9;Android;7.1.2;25);locale:" +
		"ja-jp;deviceid:ad5ad4d54477fc4f0ef9162715b530c52dd91d8d007b8865bb" +
		"6174d07c19a513;networkoperator:23430;display:Asus-user 7.1.2 2017" +
		"1130.276299 release-keys;buildid:N2G48H"
)

var (
	jar, _        = cookiejar.New(nil)
	client        = &http.Client{Transport: &Transport{}, Jar: jar}
	resolveFormat = map[int]string{
		1: "128",
		2: "320",
	}
)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", userAgent,
	)
	return http.DefaultTransport.RoundTrip(req)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	var speed int64 = 0
	n := len(p)
	wc.Downloaded += int64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	toDivideBy := time.Now().UnixMilli() - wc.StartTime
	if toDivideBy != 0 {
		speed = int64(wc.Downloaded) / toDivideBy * 1000
	}
	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
	return n, nil
}

func handleErr(errText string, err error, _panic bool) {
	errString := fmt.Sprintf("%s\n%s", errText, err)
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func parseCfg() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.Format != -1 {
		cfg.Format = args.Format
	}
	if !(cfg.Format == 1 || cfg.Format == 2) {
		return nil, errors.New("Format must be 1 or 2.")
	}
	cfg.FormatStr = resolveFormat[cfg.Format]
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "TOWER RECORDS MUSIC downloads"
	}
	if args.Lyrics {
		cfg.Lyrics = args.Lyrics
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := fmt.Sprintf("Failed to process URLs.\n%s", err)
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func makeDirs(path string) error {
	return os.MkdirAll(path, 0755)
}

func checkUrl(url string) string {
	regex := regexp.MustCompile(regexString)
	match := regex.FindStringSubmatch(url)
	if match == nil {
		return ""
	}
	return match[1]
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func sanitize(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	sanitized := regex.ReplaceAllString(filename, "_")
	return sanitized
}

func auth(email, pwd string) error {
	_url := urlBase + "login/tower"
	data := url.Values{}
	data.Set("username", email)
	data.Set("password", pwd)
	req, err := http.NewRequest(http.MethodPost, _url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", _url)
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return errors.New(do.Status)
	}
	if !strings.HasPrefix(do.Request.URL.String(), urlBase+"login/success") {
		return errors.New("Bad credentials?")
	}
	return nil
}

func getAlbumMeta(albumId string) (*AlbumMeta, error) {
	req, err := client.Get(apiBase + "album/detail/" + albumId)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return nil, errors.New(req.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(req.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	if obj.Result != 0 {
		return nil, errors.New("Bad response.")
	}
	return &obj, nil
}

func getTrackStreamUrl(_url, formatStr string) (string, string, error) {
	req, err := client.Get(_url + "/" + formatStr)
	if err != nil {
		return "", "", err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return "", "", errors.New(req.Status)
	}
	streamUrl := req.Request.URL.String()
	u, err := url.Parse(streamUrl)
	if err != nil {
		return "", "", err
	}
	queries, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", "", err
	}
	if strings.Contains(streamUrl, "/trial-music/") {
		return "", "", errors.New("The API returned a trial track stream. Expired subscription?")
	}
	retQual := queries["quality"][0]
	if retQual != formatStr {
		fmt.Println("Track unavailable in your chosen quality.")
	}
	return streamUrl, retQual, nil
}

func parseAlbumMeta(meta *AlbumMeta) map[string]string {
	parsedMeta := map[string]string{
		"album":       meta.AlbumTitle,
		"albumArtist": meta.ArtistName,
		"year":        meta.SalesDate[:4],
	}
	return parsedMeta
}

func parseTrackMeta(meta *TrackMeta, albMeta map[string]string, trackNum, trackTotal int) map[string]string {
	albMeta["artist"] = meta.ArtistName
	albMeta["title"] = meta.MusicTitle
	albMeta["track"] = strconv.Itoa(trackNum)
	albMeta["trackPad"] = fmt.Sprintf("%02d", trackNum)
	albMeta["trackTotal"] = strconv.Itoa(trackTotal)
	return albMeta
}

func downloadTrack(trackPath, url string) error {
	f, err := os.OpenFile(trackPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return errors.New(do.Status)
	}
	totalBytes := do.ContentLength
	counter := &WriteCounter{
		Total:     totalBytes,
		TotalStr:  humanize.Bytes(uint64(totalBytes)),
		StartTime: time.Now().UnixMilli(),
	}
	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	fmt.Println("")
	return err
}

func getTracktotal(meta *AlbumMeta) int {
	trackTotal := 0
	for _, disk := range meta.DiscList {
		trackTotal += len(disk.TrackList)
	}
	return trackTotal
}

func parseTemplate(templateText string, tags map[string]string) string {
	var buffer bytes.Buffer
	for {
		err := template.Must(template.New("").Parse(templateText)).Execute(&buffer, tags)
		if err == nil {
			break
		}
		fmt.Println("Failed to parse template. Default will be used instead.")
		templateText = "{{.trackPad}}. {{.title}}"
		buffer.Reset()
	}
	return buffer.String()
}

// Tracks come with metadata and covers, but no track total or year.
func writeTags(trackPath string, allTags map[string]string) error {
	tags := map[string]string{
		"tracknum": allTags["track"] + "/" + allTags["trackTotal"],
		"year":     allTags["year"],
	}
	return ap.WriteTags(trackPath, tags)
}

func getLyrics(trackId int) (*Lyrics, error) {
	query := url.Values{}
	query.Set("user-id", "")
	query.Set("type", "1")
	query.Set("access-token", accessToken)
	req, err := http.NewRequest(http.MethodGet, lyricsUrl+strconv.Itoa(trackId)+"/id-type/1", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query.Encode()
	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	var lyricsMeta LyricsMeta
	err = xml.NewDecoder(do.Body).Decode(&lyricsMeta)
	if err != nil {
		return nil, err
	}
	statusCode := lyricsMeta.Head.StatusCode
	if statusCode == "5" {
		return nil, nil
	}
	if statusCode != "0" {
		return nil, errors.New("Bad response.")
	}
	var lyrics Lyrics
	err = json.Unmarshal([]byte(lyricsMeta.Body.Data), &lyrics)
	if err != nil {
		return nil, err
	}
	return &lyrics, nil
}

func writeLyrics(lrcPath string, lyrics *Lyrics) error {
	f, err := os.OpenFile(lrcPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, lyric := range *lyrics {
		timeInt, err := strconv.ParseInt(lyric.Time, 10, 64)
		if err != nil {
			return err
		}
		ts := time.UnixMilli(timeInt).Format("04:05.00")
		line := fmt.Sprintf("[%s] %s\n", ts, lyric.Words)
		_, err = f.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	fmt.Println(`
 _____ _____ _____    ____                _           _
|_   _| __  |     |  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
  | | |    -| | | |  |  |  | . | | | |   | | . | .'| . | -_|  _|
  |_| |__|__|_|_|_|  |____/|___|_____|_|_|_|___|__,|___|___|_|
`)
}

func main() {
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		handleErr("Failed to parse config file.", err, true)
	}
	err = makeDirs(cfg.OutPath)
	if err != nil {
		handleErr("Failed to make output folder.", err, true)
	}
	err = auth(cfg.Email, cfg.Password)
	if err != nil {
		handleErr("Failed to auth.", err, true)
	}
	fmt.Println("Signed in successfully.\n")
	albumTotal := len(cfg.Urls)
	for albumNum, _url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		albumId := checkUrl(_url)
		if albumId == "" {
			fmt.Println("Invalid URL:", _url)
			continue
		}
		meta, err := getAlbumMeta(albumId)
		if err != nil {
			handleErr("Failed to get album metadata.", err, false)
			continue
		}
		parsedAlbMeta := parseAlbumMeta(meta)
		albumFolder := parsedAlbMeta["albumArtist"] + " - " + parsedAlbMeta["album"]
		fmt.Println(albumFolder)
		if len(albumFolder) > 120 {
			fmt.Println("Album folder was chopped as it exceeds 120 characters.")
			albumFolder = albumFolder[:120]
		}
		albumPath := filepath.Join(cfg.OutPath, sanitize(albumFolder))
		err = makeDirs(albumPath)
		if err != nil {
			handleErr("Failed to make album folder.", err, false)
			continue
		}
		trackNum := 0
		trackTotal := getTracktotal(meta)
		for _, disk := range meta.DiscList {
			for _, track := range disk.TrackList {
				trackNum++
				parsedMeta := parseTrackMeta(&track, parsedAlbMeta, trackNum, trackTotal)
				trackFname := parseTemplate(cfg.TrackTemplate, parsedMeta)
				sanTrackFname := sanitize(trackFname)
				trackPath := filepath.Join(albumPath, sanTrackFname+".m4a")
				exists, err := fileExists(trackPath)
				if err != nil {
					handleErr("Failed to check if track already exists locally.", err, false)
					continue
				}
				if exists {
					fmt.Println("Track already exists locally.")
					continue
				}
				streamUrl, retQual, err := getTrackStreamUrl(track.OnetimeURL, cfg.FormatStr)
				if err != nil {
					handleErr("Failed to get track stream URL.", err, false)
					continue
				}
				fmt.Printf(
					"Downloading track %d of %d: %s - AAC %s\n",
					trackNum, trackTotal, parsedMeta["title"], retQual,
				)
				err = downloadTrack(trackPath, streamUrl)
				if err != nil {
					handleErr("Failed to download track.", err, false)
					continue
				}
				err = writeTags(trackPath, parsedMeta)
				if err != nil {
					handleErr("Failed to write extra tags.", err, false)
				}
				if cfg.Lyrics {
					lyrics, err := getLyrics(track.TrackID)
					if err != nil {
						handleErr("Failed to get lyrics.", err, false)
						continue
					}
					if lyrics == nil {
						continue
					}
					lyricsPath := filepath.Join(albumPath, sanTrackFname+".lrc")
					err = writeLyrics(lyricsPath, lyrics)
					if err != nil {
						handleErr("Failed to write lyrics.", err, false)
						continue
					}
					fmt.Println("Wrote lyrics.")
				}
			}
		}
	}
}

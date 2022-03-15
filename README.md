# TOWER-RECORDS-MUSIC-Downloader
[TOWER RECORDS MUSIC](https://dereferer.me/?https://music.tower.jp/home/) downloader written in Go.
![](https://i.imgur.com/BxZEXLn.png)
[Windows, Linux and macOS binaries](https://github.com/Sorrow446/TOWER-RECORDS-MUSIC-Downloader/releases)

# Setup
Input credentials into config file.
Configure any other options if needed.
|Option|Info|
| --- | --- |
|email|Email address.
|password|Password.
|format|Download quality. 1 = AAC 128, 2 = AAC 320.
|outPath|Where to download to. Path will be made if it doesn't already exist.
|trackTemplate|Track filename naming template. Vars: album, albumArtist, artist, title, track, trackPad, trackTotal, year.
|lyrics|Get lyrics if available.

# Usage
Args take priority over the config file.    
**You can get a subscription with a foreign credit card.**

Download two albums:   
`trm_dl_x64.exe https://music.tower.jp/album/detail/1020615044 https://music.tower.jp/album/detail/1020661108`

Download a single album and from two text files:   
`trm_dl_x64.exe https://music.tower.jp/album/detail/1020615044 G:\1.txt G:\2.txt`
```
 _____ _____ _____    ____                _           _
|_   _| __  |     |  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
  | | |    -| | | |  |  |  | . | | | |   | | . | .'| . | -_|  _|
  |_| |__|__|_|_|_|  |____/|___|_____|_|_|_|___|__,|___|___|_|

Usage: trm_dl_x64.exe [--format FORMAT] [--outpath OUTPATH] [--lyrics] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --format FORMAT, -f FORMAT
                         Download quality. 1 = AAC 128, 2 = AAC 320. [default: -1]
  --outpath OUTPATH, -o OUTPATH
                         Where to download to. Path will be made if it doesn't already exist.
  --lyrics, -l           Get lyrics if available.
  --help, -h             display this help and exit
  ```
   
# Disclaimer
- I will not be responsible for how you use TOWER RECORDS MUSIC Downloader.    
- Tower Records and RecoChoku brand and names are the registered trademarks of their respective owners.    
- TOWER RECORDS MUSIC Downloader has no partnership, sponsorship or endorsement with Tower Records or RecoChoku.

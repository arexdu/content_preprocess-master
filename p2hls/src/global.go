package main

import "github.com/alecthomas/log4go"

const (
	FINISH                    int = 0
	UNKNOWN_ERR               int = -100
	PARAM_ERR                 int = -101
	INTELNAL_ERR              int = -102
	CONTENT_LIMIT_ERR         int = -1001
	CONTENT_NOT_EXIST_ERR     int = -1002
	CONTENT_ALREADY_EXIST_ERR int = -1003
	FTP_UPLOAD_ERR            int = -1005
	TIME_TIME_ERR             int = -1008
	DOWNLOAD_ERR              int = -1009
	OTHER_ERR                 int = -10000
)

var gRun int = 0
var gNow int64 = 0
var gVersion string = "1.0.0"
var gConfig Config
var gConfigPath string = "../etc/config.json"

// var gParse2hlsPath string = "../tool/parse2hls"
var gParse2hlsPath string = "/opt/fonsview/NE/parse2hls/bin/parse2hls"
var gFfmpegPath string = "../tool/ffmpeg"
var gFfprobePath string = "../tool/ffproe"
var gDownloadPath string = "/opt/fonsview/NE/parse2hls/download"

var logs = log4go.Logger{}

//var gTaskMovieMap map[string]MovieAdd
//var gWaitMovieMap map[string]MovieAdd
//var gRunMovieMap map[string]MovieAdd

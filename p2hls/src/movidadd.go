package main

import (
	"fmt"
	"net/url"
	"strings"
)

type ContentInfo struct {
	MsgType    string
	ContentId  string
	ProviderId string
}

func CreateParse2Hls(cmdStr string) int {

	cmd := StringToSlice(cmdStr, " ")
	return CmdRun(gConfig.Parse2hlsPath, cmd)
}

func CallBackMsg(movieAdd MovieAdd, ret int) {

	var cmpl MovieAddCmpl
	cmpl.uuid = movieAdd.uuid
	cmpl.MsgType = "MovieAddCmpl"
	cmpl.ContentId = movieAdd.ContentId
	cmpl.ProviderId = movieAdd.ProviderId
	cmpl.RequestId = movieAdd.RequestId
	cmpl.DstInfoList = make([]DstInfoList, 0)
	cmpl.ResultCode = ret

	fileName, _ := GetURLFileName(movieAdd.FileURL)
	var proto = strings.ToLower(GetURLProto(movieAdd.FileURL))

	if movieAdd.ContentType == 0 {
		cmpl.DownloadURL = fmt.Sprintf("%s%s", movieAdd.FileDstURL, fileName)
		logs.Info("[%s]DownloadURL:%s", movieAdd.uuid, cmpl.DownloadURL)
	} else if movieAdd.ContentType == 1 {

		if proto == "ftp" {
			cmpl.DownloadURL = fmt.Sprintf("%s%s", movieAdd.FileDstURL, "main.m3u8")
		} else {
			cmpl.DownloadURL = fmt.Sprintf("http://%s:%d/%s_%s/main.m3u8", gConfig.Host, gConfig.Port, movieAdd.ProviderId, movieAdd.ContentId)
		}

	} else if movieAdd.ContentType == 2 {
		cmpl.DownloadURL = fmt.Sprintf("%s%s", movieAdd.FileDstURL, fileName)
	}

	logs.Info("[%s]MovieAddCmpl:%v", movieAdd.uuid, cmpl)
	HttpMovieAddCmpl(gConfig.CallBack, cmpl)
}

func DownLoadFile(movieAdd MovieAdd, filename string) int {

	var ret int
	var proto = strings.ToLower(GetURLProto(movieAdd.FileURL))
	//filename, _ := GetURLFileName(movieAdd.FileURL)
	//stroePath := movieAdd.storeDir + filename

	logs.Info("[%s]Parse2Hls runing download proto :%s,url:%s", movieAdd.uuid, proto, movieAdd.FileURL)
	switch proto {
	case "ftp":
		u, err := url.Parse(movieAdd.FileURL)
		if err != nil {
			logs.Error("[%s]ftp url [%s] Parse error :%s", movieAdd.uuid, movieAdd.FileURL, err.Error())
			//CallBackMsg(movieAdd, PARAM_ERR)
			return PARAM_ERR
		}

		var ftpconn FTPInfo
		ftpconn.uuid = movieAdd.uuid
		ftpconn.Host = u.Host
		ftpconn.DownloadFile = u.Path
		ftpconn.StroePath = movieAdd.storeDir
		ftpconn.PassWord, _ = u.User.Password()
		ftpconn.User = u.User.Username()

		ret = DownLoadFileFromFtp(ftpconn)
		break
	case "http":
		ret = HttpDownloadFile(movieAdd.FileURL, movieAdd.storeDir+filename, movieAdd.uuid)
		break
	default:
		logs.Error("[%s]no found proto,url:[%s]", movieAdd.uuid, movieAdd.FileURL)
		ret = DOWNLOAD_ERR
	}
	return ret
}

func Parse2Hls(movieAdd MovieAdd) {

	logs.Info("[%s]Parse2Hls start !", movieAdd.uuid)

	// 检查下载目录是否存在，不存在就创建目录
	if !CheckDirExist(movieAdd.storeDir) {
		CreateDir(movieAdd.storeDir)
		logs.Debug("[%s]CreateDir:%s", movieAdd.uuid, movieAdd.storeDir)
	}

	// 判断协议是ftp还是http下载文件到指定目录
	filename, _ := GetURLFileName(movieAdd.FileURL)
	ret := DownLoadFile(movieAdd, filename)
	if ret < 0 {
		CallBackMsg(movieAdd, ret)
		return
	}

	distDirCmd := "-F " + movieAdd.storeDir + "/"
	stroePathCmd := gConfig.Parse2HlsParam + distDirCmd + " -u " + movieAdd.storeDir + filename

	logs.Info("[%s]cmd:parse2hls %s", movieAdd.uuid, stroePathCmd)
	ret = CreateParse2Hls(stroePathCmd)

	if len(movieAdd.FileDstURL) > 0 {
		u, err := url.Parse(movieAdd.FileDstURL)
		if err != nil {
			logs.Error("[%s]ftp url [%s] Parse error :%s", movieAdd.uuid, movieAdd.FileURL, err.Error())
			CallBackMsg(movieAdd, FTP_UPLOAD_ERR)
			return
		}

		var ftpconn FTPInfo
		ftpconn.uuid = movieAdd.uuid
		ftpconn.Host = u.Host
		ftpconn.PassWord, _ = u.User.Password()
		ftpconn.User = u.User.Username()
		ftpconn.UploadDir = u.Path
		ftpconn.StroePath = movieAdd.storeDir
		ftpconn.FileNameList = append(ftpconn.FileNameList, "main.m3u8")
		ftpconn.FileNameList = append(ftpconn.FileNameList, "0.m3u8")
		ftpconn.FileNameList = append(ftpconn.FileNameList, "0_iframe.m3u8")

		tsList, err := getTsFileList(movieAdd.storeDir + "/" + "0.m3u8")
		for i := 0; i < len(tsList); i++ {
			ftpconn.FileNameList = append(ftpconn.FileNameList, tsList[i])
			logs.Info("[%s]path:%s,tsName:%s", ftpconn.uuid, movieAdd.storeDir, tsList[i])
		}

		ret = UploadFileToFtp(ftpconn)
		if ret != FINISH {
			CallBackMsg(movieAdd, FTP_UPLOAD_ERR)
			return
		}
	}

	CallBackMsg(movieAdd, ret)
	logs.Info("[%s]Parse2Hls finish !", movieAdd.uuid)
	//var dstInfo DstInfoList
	//dstInfo.DstId =
}

// 直播切片有drm加密，点播切片好像没看到
func Hls2HlsEncrypt(movieAdd MovieAdd) {

}

func FileUpload2FTP(movieAdd MovieAdd) {

	logs.Info("[%s]FileUpload2FTP start !", movieAdd.uuid)

	// 检查下载目录是否存在，不存在就创建目录
	if !CheckDirExist(movieAdd.storeDir) {
		CreateDir(movieAdd.storeDir)
		logs.Debug("[%s]CreateDir:%s", movieAdd.uuid, movieAdd.storeDir)
	}

	// 判断协议是ftp还是http下载文件到指定目录
	filename, _ := GetURLFileName(movieAdd.FileURL)
	ret := DownLoadFile(movieAdd, filename)
	if ret < 0 {
		CallBackMsg(movieAdd, ret)
		return
	}

	/**/
	u, err := url.Parse(movieAdd.FileDstURL)
	if err != nil {
		logs.Error("[%s]ftp url [%s] Parse error :%s", movieAdd.uuid, movieAdd.FileDstURL, err.Error())
		CallBackMsg(movieAdd, PARAM_ERR)
		return
	}

	idx := strings.LastIndex(movieAdd.FileURL, "/")
	fileName := movieAdd.FileURL[idx+1:]
	var ftpconn FTPInfo
	ftpconn.uuid = movieAdd.uuid
	ftpconn.Host = u.Host
	ftpconn.PassWord, _ = u.User.Password()
	ftpconn.User = u.User.Username()
	ftpconn.UploadDir = u.Path
	ftpconn.StroePath = movieAdd.storeDir
	ftpconn.FileNameList = append(ftpconn.FileNameList, fileName)

	ret = UploadFileToFtp(ftpconn)
	CallBackMsg(movieAdd, ret)
	logs.Info("[%s]FileUpload2FTP finish !", movieAdd.uuid)

}

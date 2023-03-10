package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	//logs "github.com/alecthomas/log4go"
)

func MovieAddFunc(w http.ResponseWriter, body []byte) {

	/*
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logs.Error("MovieAddFunc revc body error:" + err.Error())
			fmt.Fprintf(w, ResponseMsgStr(PARAM_ERR))
			return
		}
	*/

	var movieAdd MovieAdd
	err := json.Unmarshal(body, &movieAdd)
	if err != nil {
		logs.Error("[%s]parse json error:%s,msg:%s", movieAdd.uuid, err.Error(), string(body))
		fmt.Fprintf(w, ResponseMsgStr(PARAM_ERR))
		return
	}

	movieAdd.uuid = GetUUID()
	movieAdd.storeDir = gConfig.DownLoadDir + movieAdd.ProviderId + "_" + movieAdd.ContentId
	//gTaskMovieMap[movieAdd.uuid] = movieAdd
	if movieAdd.ContentType == 0 {
		// TODO 测试使用
		//go Parse2Hls(movieAdd)
		// 0:hls->hls（hls内容加密） 好像不需要

		go FileUpload2FTP(movieAdd)
		go Hls2HlsEncrypt(movieAdd)
	} else if movieAdd.ContentType == 1 {
		// 1:大ts->hls (大ts文件切片)
		go Parse2Hls(movieAdd)
	} else if movieAdd.ContentType == 2 {
		// 2大文件->大文件(大文件下载后上传FTP)
		go FileUpload2FTP(movieAdd)
	} else {
		logs.Error("[%s] movieAdd ContentType Unknow [%d]", movieAdd.uuid, movieAdd.ContentId)
	}

	logs.Info("[%s]MovieAdd run:%s", movieAdd.uuid, string(body))
	fmt.Fprintf(w, ResponseMsgStr(FINISH))
}

// 删除一个处理后的vod
func MovieAddDelFunc(w http.ResponseWriter, body []byte) {

	/*
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logs.Error("MovieAddDelFunc revc body error:" + err.Error())
			fmt.Fprintf(w, ResponseMsgStr(PARAM_ERR))
			return
		}
	*/

	var movidDel MovieDel
	movidDel.uuid = GetUUID()
	err := json.Unmarshal(body, &movidDel)
	if err != nil {
		logs.Error("[%s]parse json error:%s,msg:%s", movidDel.uuid, err.Error(), string(body))
		fmt.Fprintf(w, ResponseMsgStr(PARAM_ERR))
		return
	}
	logs.Info("[%s]MovieAddDel:%s", movidDel.uuid, string(body))

	cmdStr := "-rf " + gConfig.DownLoadDir + movidDel.ProviderId + "_" + movidDel.ContentId
	cmd := StringToSlice(cmdStr, " ")
	CmdRun("/bin/rm", cmd)
	logs.Info("[%s]cmd /bin/rm %s", movidDel.uuid, cmd)
	fmt.Fprintf(w, ResponseMsgStr(FINISH))
}

// 反馈一个vod处理结果
func HttpMovieAddCmpl(url string, movieInfo MovieAddCmpl) {

	msg, _ := json.Marshal(movieInfo)
	body := string(msg)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		logs.Error("[%s]create post request error:%s", movieInfo.uuid, err.Error())
	}
	logs.Debug("[%s]MovieAddCmpl request msg:%s", movieInfo.uuid, body)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uuid", movieInfo.uuid)
	//req.Header.Set("Accept-Encoding","none")
	//req.Header.Set("Accept-Encoding", "identity")

	resp, err := client.Do(req)
	if err != nil {
		logs.Error("[%s]MovieAddCmpl client error,msg:%s", movieInfo.uuid, err.Error())
		return
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("[%s]MovieAddCmpl request error,msg:%s", movieInfo.uuid, err.Error())
	}

	logs.Info("[%s]MovieAddCmpl request finish,res:%s", movieInfo.uuid, string(res))
}

// 提供http下载服务
func HttpServiceDownload(w http.ResponseWriter, r *http.Request) {
	uuid := GetUUID()
	logs.Info("[%s][%s]HOST:%s%s,len:%d", uuid, r.RemoteAddr, r.Host, r.RequestURI, r.ContentLength)
	uri := gDownloadPath + r.RequestURI

	f, err := os.Open(uri)
	if err != nil {
		logs.Error("[%s]download msg:%s", uuid, err.Error())
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, ResponseMsgStr(CONTENT_NOT_EXIST_ERR))
		return
	}

	logs.Info("[%s]read file:%s", uuid, uri)
	defer f.Close()
	_, err = io.Copy(w, f)

	//file, _ := ioutil.ReadFile(uri)
	//res, err := w.Write(file)
	//logs.Info("[%s]HttpService finish,res,msg:%s", uuid, err.Error())
}

// 处理任务
func MovieTaskFunc(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logs.Error("MovieTaskFunc revc body error:" + err.Error())
		fmt.Fprintf(w, ResponseMsgStr(PARAM_ERR))
		return
	}

	msgType := GetMsgType(body)
	switch msgType {
	case "MovieAdd":
		MovieAddFunc(w, body)
		break
	case "MovieDel":
		MovieAddDelFunc(w, body)
		break
	default:
		logs.Error("MsgType no found ! res:%s", string(body))
		fmt.Fprintf(w, ResponseMsgStr(UNKNOWN_ERR))
		return
	}
}

// 提供http下载服务
func HttpService(w http.ResponseWriter, r *http.Request) {
	logs.Info("[%s]HOST:%s%s,len:%d", r.RemoteAddr, r.Host, r.RequestURI, r.ContentLength)
	switch r.Method {
	case "POST":
		MovieTaskFunc(w, r)
	case "GET":
		HttpServiceDownload(w, r)
	}
}

// http文件下载
func HttpDownloadFile(url string, path string, uuid string) int {
	resp, err := http.Get(url)
	if err != nil {
		logs.Info("[%s] http Get %s,msg:%s", uuid, url, err.Error())
		return -1
	}
	defer resp.Body.Close()

	// 创建文件
	out, err := os.Create(path)
	if err != nil {
		logs.Info("[%s] os Create %s,msg:%s", uuid, path, err.Error())
		return -1
	}
	defer out.Close()

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logs.Info("[%s] io Copy,msg:%s", uuid, err.Error())
		return -1
	}

	return 0
}

// 初始化http服务
func InitHttpService(host string, port int) {

	//http.HandleFunc("/MovieAdd", MovieAddFunc)
	//http.HandleFunc("/MovieAddDel", MovieAddDelFunc)
	http.HandleFunc("/", HttpService)

	// 启动http服务
	service := fmt.Sprintf("%s:%d", "0.0.0.0", port)
	go http.ListenAndServe(service, nil)

	if host != "0.0.0.0" {
		service = fmt.Sprintf("%s:%d", host, port)
		go http.ListenAndServe(service, nil)
	}

	logs.Info("Listen Service %s", service)
}

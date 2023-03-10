package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/bitly/go-simplejson"
	//uuid "github.com/satori/go.uuid"
	uuid "github.com/google/uuid"
)

type MovieAdd struct {
	MsgType     string
	RequestId   int
	ContentId   string
	ProviderId  string
	ContentType int
	FileURL     string
	DrmFlag     int
	FileDstURL  string

	uuid     string
	storeDir string
}

type MovieDel struct {
	MsgType    string
	ContentId  string
	ProviderId string

	uuid string
}

type DstInfoList struct {
	DstId      int
	BandWidths []string
	DstUrl     string
}

type MovieAddCmpl struct {
	MsgType     string
	RequestId   int
	ContentId   string
	ProviderId  string
	ResultCode  int
	DownloadURL string

	DstInfoList []DstInfoList

	uuid string
}

type Response struct {
	ResultCode int
}

func CmdRun(app string, params []string) int {
	cmd := exec.Command(app, params...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logs.Error("[%s]cmd run error! cmd:%s,msg:%s", app, cmd, err.Error())
		return INTELNAL_ERR
	}
	//fmt.Printf("command output: %q", out.String())
	return FINISH
}

func StringToSlice(str string, sep string) []string {
	return strings.Split(str, sep)
}

func ResponseMsgStr(codeID int) string {
	msg := fmt.Sprintf("{\"ResultCode\": %d}", codeID)
	return msg
}

func GetUUID() string {

	u1 := uuid.New()
	return fmt.Sprintf("%d", u1.ID())

	//u2, _ := uuid.NewRandom()
	//id := uuid.NewV4().Bytes()
	//return base32.StdEncoding.EncodeToString(id)
}

func GetURLProto(url string) string {
	var arr = strings.Split(url, "://")
	return arr[0]
}
func GetURLFileName(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	idx := strings.LastIndex(u.Path, "/")
	return u.Path[idx:], err
}

// 检查目录是否存在
func CheckDirExist(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return true
}

// 创建文件夹
func CreateDir(dir string) error {
	err := os.Mkdir(dir, os.ModePerm)
	return err
}

// 删除文件夹
func DeleteDir(dir string) error {
	err := os.RemoveAll(dir)
	return err
}

func GetMsgType(body []byte) string {
	js, err := simplejson.NewJson(body)
	if err != nil {
		// json解析失败
		logs.Error("GetMsgType parse json error !res:%s", string(body))
		return ""
	}

	var str string
	if value, ok := js.CheckGet("MsgType"); ok {
		str, _ = value.String()
	}
	return str
}

func getTsFileList(filePath string) ([]string, error) {

	var tsList = make([]string, 0)
	fileHanle, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	defer fileHanle.Close()
	reader := bufio.NewReader(fileHanle)

	// 按行处理txt
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		str := string(line)
		if strings.LastIndex(str, ".ts") >= 0 {
			tsList = append(tsList, str)
		}
	}
	return tsList, nil
}

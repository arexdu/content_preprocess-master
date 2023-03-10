package main

import (
	"bytes"
	"io/ioutil"
	"strings"

	//logs "github.com/alecthomas/log4go"
	"github.com/jlaffaye/ftp"
)

type FTPInfo struct {
	Host         string
	User         string
	PassWord     string
	DownloadFile string
	StroePath    string

	FileNameList []string
	UploadDir    string

	uuid string
}

// FTP下载函数
func DownLoadFileFromFtp(ftpconn FTPInfo) int {
	host := ftpconn.Host
	user := ftpconn.User
	password := ftpconn.PassWord
	filePath := ftpconn.DownloadFile

	// 连接到FTP服务器
	c, err := ftp.Dial(host)
	if err != nil {
		logs.Error("[%s]ftp Connect %s error", ftpconn.uuid, filePath, err.Error())
		return DOWNLOAD_ERR
	}

	defer c.Quit()

	// 用户名和密码登陆ftp
	if err := c.Login(user, password); err != nil {
		logs.Error("[%s]ftp Login %s error", ftpconn.uuid, filePath, err.Error())
		return DOWNLOAD_ERR
	}

	/*
		    // 获得当前服务器根目录地址
			dir, err := c.CurrentDir()
			if err != nil {
				logs.Error("[%s]ftp CurrentDir msg:%s", ftpconn.uuid, err.Error())
				return DOWNLOAD_ERR
			}
			logs.Info("ftp dir:%s", dir)
	*/

	idx := strings.LastIndex(filePath, "/")

	var fileDir string
	if idx > 0 {
		fileDir = filePath[1:idx]
		//进入文件夹，可以嵌套 dir1/dir11
		err = c.ChangeDir(fileDir)
		if err != nil {
			logs.Error("[%s]ftp ChangeDir :%s,msg:%s", ftpconn.uuid, fileDir, err.Error())
			return DOWNLOAD_ERR
		}
	}
	fileName := filePath[idx+1:]

	logs.Info("[%s]ftp download dir:%s,fileName:%s", ftpconn.uuid, fileDir, fileName)

	// 返回上一级
	//err = c.ChangeDirToParent()

	// 获得需要下载的文件数据
	r, err := c.Retr(fileName)
	if err != nil {
		logs.Error("[%s]ftp Retr %s error:%s", ftpconn.uuid, filePath, err.Error())
		return DOWNLOAD_ERR
	}
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		logs.Error("[%s]ioutil.ReadAll %s error :%s", ftpconn.uuid, filePath, err.Error())
		return DOWNLOAD_ERR
	}

	path := ftpconn.StroePath + "/" + fileName
	err = ioutil.WriteFile(path, buf, 0666)
	if err != nil {
		logs.Error("[%s]ioutil.WriteFile %s error :%s", ftpconn.uuid, filePath, err.Error())
		return DOWNLOAD_ERR
	}

	logs.Info("[%s]FTP download %s -> %s successful!", ftpconn.uuid, filePath, path)
	return FINISH
}

/*
func main() {
	DownLoadFileFromFtp()
}
*/

func EnsureFtpDirExist(c *ftp.ServerConn, dir string) error {
	// 这里不能直接 MakeDir，有权限问题
	_, err := c.List(dir)
	if err != nil {
		logs.Error("ftp dir list err:%s", dir)
	}
	if er := c.MakeDir(dir); er != nil {
		if er.Error() == "550 Directory with same name already exists." {
			return nil
		}
		return er
	}
	return nil
}

func CreateMutilDir(c *ftp.ServerConn, path string, uuid string) error {

	dirList := strings.Split(path, "/")
	var dir string
	for i := 0; i < len(dirList); i++ {
		if i > 0 {
			dir += "/"
		}
		dir += dirList[i]
		err := c.MakeDir(dir)
		if err != nil {
			// 创建的文件夹已存在时 不做处理
			if strings.Index(err.Error(), "550") < 0 {
				logs.Error("[%s]ftp MakeDir err:[%s],msg:%s", uuid, dir, err.Error())
				return err
			}
			logs.Warn("[%s]ftp MakeDir :[%s],msg:%s", uuid, dir, err.Error())
		}
	}

	return nil
}

func UploadFileToFtp(ftpconn FTPInfo) int {

	host := ftpconn.Host
	user := ftpconn.User
	password := ftpconn.PassWord
	FileNameList := ftpconn.FileNameList
	uploadDir := ftpconn.UploadDir[1:]

	// 去掉目录最后一个反斜杠
	if uploadDir[len(uploadDir)-1:] == "/" {
		uploadDir = uploadDir[:len(uploadDir)-1]
	}

	path := ftpconn.StroePath

	// 连接到FTP服务器
	c, err := ftp.Dial(host)
	if err != nil {
		logs.Error("[%s]ftp Connect %s error", ftpconn.uuid, path, err.Error())
		return FTP_UPLOAD_ERR
	}

	defer c.Quit()

	// 用户名和密码登陆ftp
	if err := c.Login(user, password); err != nil {
		logs.Error("[%s]ftp Login %s error", ftpconn.uuid, path, err.Error())
		return FTP_UPLOAD_ERR
	}

	//err = EnsureFtpDirExist(c, uploadDir)
	/*
		err = c.MakeDir(uploadDir)
		if err != nil {
			if strings.Index(err.Error(), "550") < 0 {
				logs.Error("[%s]ftp MakeDir :%s,msg:%s", ftpconn.uuid, uploadDir, err.Error())
				return FTP_UPLOAD_ERR
			}
			logs.Warn("[%s]ftp MakeDir :%s,msg:%s", ftpconn.uuid, uploadDir, err.Error())
		}
	*/

	err = CreateMutilDir(c, uploadDir, ftpconn.uuid)
	if err != nil {
		logs.Error("[%s]ftp CreateMutilDir msg:%s", ftpconn.uuid, err.Error())
		return FTP_UPLOAD_ERR
	}
	dir, err := c.CurrentDir()
	if err != nil {
		logs.Error("[%s]ftp CurrentDir msg:%s", ftpconn.uuid, err.Error())
		return FTP_UPLOAD_ERR
	}

	logs.Info("[%s]current dir :%s,upload:%s", ftpconn.uuid, dir, uploadDir)

	err = c.ChangeDir(uploadDir)
	if err != nil {
		logs.Error("[%s]ftp ChangeDir :%s,msg:%s", ftpconn.uuid, uploadDir, err.Error())
		return FTP_UPLOAD_ERR
	}

	for i := 0; i < len(FileNameList); i++ {
		filePath := FileNameList[i]
		//file, _ := os.Open(filePath)
		//defer file.Close()
		buf, _ := ioutil.ReadFile(filePath)
		err = c.Stor(FileNameList[i], bytes.NewBuffer(buf))
		if err != nil {
			logs.Error("[%s]ftp Stor :%s,msg:%s", ftpconn.uuid, FileNameList[i], err.Error())
			return FTP_UPLOAD_ERR
		}
		logs.Info("[%s]ftp upload:%s finish !", ftpconn.uuid, uploadDir+FileNameList[i])
	}

	//c.Logout()
	return FINISH
}

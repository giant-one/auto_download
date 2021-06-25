package main

import (
	"../unit"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/cheggaaa/pb"
	"github.com/zxysilent/logs"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var localFiles = []string{""}
var token string

func main()  {
	// 使用默认实例
	// 退出时调用，确保日志写入文件中
	defer logs.Flush()
	// 设置日志输出等级
	// 开发环境下设置输出等级为DEBUG，线上环境设置为INFO
	logs.SetLevel(logs.DEBUG)
	// 设置输出调用信息
	logs.SetCallInfo(true)
	// 设置同时显示到控制台
	// 默认只输出到文件
	logs.SetConsole(true)
	logs.Info("-----sync begin-----")
	//获取当前路径
	path, err := GetCurrentPath()
	logs.Info("本地路径:"+path)
	if err != nil {
		logs.Error(err)
	}
	//获取token
	token = getToken(path)
	if token == "" {
		logs.Error("token is empty")
		os.Exit(3)
	}
	logs.Info("token:", token)
	// 加密密钥
	passphrase := "111111111111111111111111"
	company := unit.Decrypt(token, passphrase)
	localFilesPath := path + "\\"+ company
	logs.Info("本地课程路径:"+localFilesPath)
	//读取本地目录
	getFiles(localFilesPath)
	logs.Infof("本地文件 %s", localFiles)
	//获取服务器目录
	remoteFileList := getRemoteFilesInfo()
	logs.Infof("服务器文件%v", remoteFileList)
	if len(remoteFileList) < 0 {
		logs.Error("远程文件获取失败")
		os.Exit(3)
	}
	localFiles = handleFiles(localFiles)
	logs.Infof("处理后本地文件%v", localFiles)
	downloadFiles := difference(remoteFileList, localFiles)
	logs.Infof("需要下载文件%v", downloadFiles)
	deleteFiles := difference(localFiles, remoteFileList)
	logs.Infof("需要删除文件%v", deleteFiles)
	//下载文件
	for _, file := range downloadFiles {
		downloadFile(file, path + "\\" + file)
	}
	//删文件
	for _, file := range deleteFiles {
		if file == "" {
			continue
		}
		deleteFile(path + "\\" + file)
	}
	fmt.Println("服务器的：", remoteFileList)
	fmt.Println("要下载的：", downloadFiles)
	logs.Info("-----sync end-----")
}

func getToken(path string) string {
	bytes,err := ioutil.ReadFile(path + "conf.ini")
	if err != nil {
		logs.Error(err)
	}
	fmt.Println("total bytes read：",len(bytes))
	fmt.Println("string read:",string(bytes))
	return string(bytes)
}

func deleteFile(filepath string)  {
	err := os.Remove(filepath)
	if err != nil {
		// 删除失败
		logs.Error(err)
	}
	// 删除成功
	logs.Info("file delete success", filepath)
}

func handleFiles(localFiles []string) []string {
	res := make([]string, len(localFiles) -1)
	for _, file := range localFiles {
		if file == "" {
			continue
		}
		str1 := strings.Split(file, "\\")
		res = append(res, strings.TrimSpace(str1[1]))
	}
	return res
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		path = strings.Replace(path, "\\", "/", -1)
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "", errors.New(`Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

func downloadFile(fileName string, saveName string) {

	escapeUrl := url.QueryEscape(fileName)
	_, showName := filepath.Split(fileName)
	savePath, _ := filepath.Split(saveName)

	sourceName, destName := "http://192.168.1.104/index.php?method=download_file&fileName="+escapeUrl, saveName

	// check source
	var source io.Reader
	var sourceSize int64
	if strings.HasPrefix(sourceName, "http://") {
		// open as url
		resp, err := http.Get(sourceName)
		if err != nil {
			logs.Error("Can't get %s: %v\n", sourceName, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logs.Error("Server return non-200 status: %v\n", resp.Status)
			return
		}
		i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		sourceSize = int64(i)
		source = resp.Body
	} else {
		// open as file
		s, err := os.Open(sourceName)
		if err != nil {
			fmt.Printf("Can't open %s: %v\n", sourceName, err)
			return
		}
		defer s.Close()
		// get source size
		sourceStat, err := s.Stat()
		if err != nil {
			fmt.Printf("Can't stat %s: %v\n", sourceName, err)
			return
		}
		sourceSize = sourceStat.Size()
		source = s
	}

	if !IsDir(savePath) {
		createDir(savePath)
	}

	// create dest
	dest, err := os.Create(destName)
	if err != nil {
		logs.Error("Can't create %s: %v\n", fileName, err)
		return
	}
	defer dest.Close()
	// create bar
	bar := pb.New(int(sourceSize)).Postfix(showName).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.Start()

	// create proxy reader
	reader := bar.NewProxyReader(source)

	// and copy from reader
	io.Copy(dest, reader)
	bar.Finish()
	logs.Info("download success", fileName)
}

func IsDir(fileAddr string)bool{
	s,err:=os.Stat(fileAddr)
	if err!=nil{
		return false
	}
	return s.IsDir()
}

func createDir(dirName string) bool {
	err :=os.Mkdir(dirName,755)
	if err!=nil{
		return false
	}
	return true
}

func getRemoteFilesInfo() []string {
	r, err := http.Get("http://192.168.1.104/index.php?method=get_file_info&token="+token)
	if err != nil {
		panic(err)
	}
	defer func() {_ = r.Body.Close()}()

	content, err := ioutil.ReadAll(r.Body)
	logs.Info("get_remote_file_info", string(content))
	if err != nil {
		panic(err)
	}
	result, err := simplejson.NewJson([]byte(content))
	if err != nil {
		logs.Error("get_remote_file_info_err", err)
	}
	fileList, err := result.Get("data").StringArray()
	return fileList
}

func getFiles(folder string){
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			getFiles(folder + "/" + file.Name())
		} else {
			localFiles = append(localFiles, folder + "/" + file.Name())
		}
	}
}


func difference(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := intersect(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}

	for _, value := range slice1 {
		times, _ := m[value]
		if times == 0 {
			nn = append(nn, value)
		}
	}
	return nn
}

//求交集
func intersect(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}
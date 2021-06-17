package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

var localFilesPath = []string{""}

func main()  {
	//读取本地目录
	getFiles("G:\\课程")
	//获取服务器目录
	remoteFileList := getRemoteFilesInfo()
	//fmt.Printf("%v", remoteFileList)
	localFilesPath = handleFilePath(localFilesPath)
	fmt.Printf("%v", localFilesPath)
	di := difference(localFilesPath, remoteFileList)

	//下载文件

	fmt.Println("slice1与slice2的差集为：", di)
}

func handleFilePath(localFilesPath []string) []string {
	for i:= 0;i<len(localFilesPath);i++{
		//fmt.Println(localFilesPath[i])
		localFilesPath[i] = split(localFilesPath[i])
	}
	return localFilesPath
}

func split(str string) string {
	//lens := len(str)
	//str = string([]rune(str)[3:25])
	//str = strings.Replace(str, " ", "", -1)
	str = string([]rune(str)[3:20])
	return str
}

func getRemoteFilesInfo() []string {
	r, err := http.Get("http://192.168.1.6?method=get_file_info")
	if err != nil {
		panic(err)
	}
	defer func() {_ = r.Body.Close()}()

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	result, err := simplejson.NewJson([]byte(content))
	if err != nil {
		fmt.Printf("%v\n", err)
		//return make([err]string, 1)
	}
	fileList, err := result.Get("data").StringArray()
	return fileList
}

func downloadFile(fileName string)  {
	request, err := http.NewRequest(http.MethodGet,"http://192.168.1.6", nil)
	if err != nil {
		panic(err)
	}

	params := make(url.Values)
	params.Add("method", "download_file")
	params.Add("fileName", fileName)
	request.URL.RawQuery = params.Encode()
	r, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	defer func() {_ = r.Body.Close()}()

	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer func() {_= f.Close()}()
	n, err := io.Copy(f, r.Body)
	fmt.Println(n, err)
}

func getFiles(folder string){
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			getFiles(folder + "/" + file.Name())
		} else {
			//fmt.Println(folder + "/" + file.Name())
			localFilesPath = append(localFilesPath, folder + "/" + file.Name())
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
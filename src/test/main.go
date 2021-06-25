package main

import (
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	srt := "D:/project/golang/auto_downlaod/src/main/\\课程/001 Swoole和TCP三次握手.mp4"
	str1 := strings.Split(srt, "\\")
	fmt.Printf("%v", str1)
	//ExampleCopy()
	//Example_multiple()
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
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
	//fmt.Println("path111:", path)
	if runtime.GOOS == "windows" {
		path = strings.Replace(path, "\\", "/", -1)
	}
	//fmt.Println("path222:", path)
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "", errors.New(`Can't find "/" or "\".`)
	}
	//fmt.Println("path333:", path)
	return string(path[0 : i+1]), nil
}

func Example_multiple() {
	// create bars
	first := pb.New(200).Prefix("First ")
	second := pb.New(200).Prefix("Second ")
	third := pb.New(200).Prefix("Third ")
	// start pool
	pool, err := pb.StartPool(first, second, third)
	if err != nil {
		panic(err)
	}
	// update bars
	wg := new(sync.WaitGroup)
	for _, bar := range []*pb.ProgressBar{first, second, third} {
		wg.Add(1)
		go func(cb *pb.ProgressBar) {
			for n := 0; n < 200; n++ {
				cb.Increment()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
			cb.Finish()
			wg.Done()
		}(bar)
	}
	wg.Wait()
	// close pool
	pool.Stop()
}

func ExampleCopy() {
	// check args
	//if len(os.Args) < 3 {
	//	printUsage()
	//	return
	//}
	//sourceName, destName := os.Args[1], os.Args[2]
	sourceName, destName := "http://192.168.1.104/index.php?method=download_file&fileName=%E5%8F%91%E5%B8%83%E8%A7%86%E9%A2%91%2F001%20Swoole%E5%92%8CTCP%E4%B8%89%E6%AC%A1%E6%8F%A1%E6%89%8B.mp4", "a.mp4"

	// check source
	var source io.Reader
	var sourceSize int64
	if strings.HasPrefix(sourceName, "http://") {
		// open as url
		resp, err := http.Get(sourceName)
		if err != nil {
			fmt.Printf("Can't get %s: %v\n", sourceName, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Server return non-200 status: %v\n", resp.Status)
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

	// create dest
	dest, err := os.Create(destName)
	if err != nil {
		fmt.Printf("Can't create %s: %v\n", destName, err)
		return
	}
	defer dest.Close()
	fmt.Print("a.mp4")
	// create bar
	bar := pb.New(int(sourceSize)).Postfix(" a.mp4 ").SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.Start()

	// create proxy reader
	reader := bar.NewProxyReader(source)

	// and copy from reader
	io.Copy(dest, reader)
	bar.Finish()
}

func printUsage() {
	fmt.Println("copy [source file or url] [dest file]")
}
package main

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	ExampleCopy()
	//Example_multiple()
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
	sourceName, destName := "http://vd3.bdstatic.com/mda-kgvf0kqtznpgdig6/v1-cae/sc/mda-kgvf0kqtznpgdig6.mp4?v_from_s=tc_haokan_4469&auth_key=1623928979-0-0-4eb57fca164ea6036d8ce3e0d2ee931f&bcevod_channel=searchbox_feed&pd=1&pt=3&abtest=3000156_2", "a.mp4"

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
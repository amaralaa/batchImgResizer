package main

import (
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

func main() {
	var imgList []string
	var wg sync.WaitGroup
	maxGoroutines, e := strconv.Atoi(os.Args[1])
	if e != nil {
		log.Fatalln("Please insert number of threads as intger 1 to 100")
	}
	if maxGoroutines < 1 {
		log.Fatalln("Please insert number of threads as intger 1 to 100")
	}
	guard := make(chan struct{}, maxGoroutines)
	log.Println("Start Reading Dir.")
	err := filepath.Walk("resize", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			path = strings.Replace(path, "resize", "", 1)
			err := os.MkdirAll("new"+path, 0600)
			if err != nil {
				return err
			}
		}
		if strings.ToLower(filepath.Ext(path)) == ".jpg" || strings.ToLower(filepath.Ext(path)) == ".jpeg" {
			imgList = append(imgList, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Finish Creating New Dir.")
	log.Printf("Start resizing %d Images\n", len(imgList))
	for _, img := range imgList {
		wg.Add(1)
		guard <- struct{}{} // would block if guard channel is already filled
		go func(p string) {
			resizeImg(p)
			wg.Done()
			<-guard
		}(img)
	}
	wg.Wait()
	log.Println("Done resizing images")
}

func resizeImg(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatalln(err)
	}
	file.Close()
	img = resize.Resize(3072, 0, img, resize.Lanczos3)
	out, err := os.Create(strings.Replace(path, "resize", "new", 1))
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()
	jpeg.Encode(out, img, nil)
	// log.Printf("img %s finished", path)
}

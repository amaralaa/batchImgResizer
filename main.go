package main

import (
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
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
	folderList := readList()
	sort.StringSlice(folderList).Sort()
	filepath.Walk("resize", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			for _, f := range folderList {
				if f == info.Name() {
					log.Println("Found : ", f)
					new := makeList(path)
					imgList = append(imgList, new...)
				}
			}
		}
		return nil
	})
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
		log.Println(err)
		return
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Println(err)
		return
	}
	file.Close()
	img = resize.Resize(3072, 0, img, resize.Lanczos3)
	out, err := os.Create(strings.Replace(path, "resize", "new", 1))
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()
	jpeg.Encode(out, img, nil)
	// log.Printf("img %s finished", path)
}

func readList() []string {
	data, err := ioutil.ReadFile("list.txt")
	if err != nil {
		log.Fatalln("failed reading list.txt", err)
	}
	lines := strings.Split(string(data), "\n")
	return lines
}

func makeList(folder string) []string {
	var imgList []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			path = strings.Replace(path, "resize", "", 1)
			err := os.MkdirAll("new"+path, 0644)
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
	return imgList
}

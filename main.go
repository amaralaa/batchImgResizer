package main

import (
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
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
			resizeImgOld(p)
			wg.Done()
			<-guard
		}(img)
	}
	wg.Wait()
	log.Println("Done resizing images")
}

func resizeImg(path string, done *sync.WaitGroup) {
	file, err := os.Open(path)
	if err != nil {
		done.Done()
		log.Fatalln(err)
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		done.Done()
		log.Fatalln(err)
	}
	file.Close()
	m := resize.Resize(3072, 0, img, resize.Lanczos3)
	out, err := os.Create(strings.Replace(path, "resize", "new", 1))
	if err != nil {
		done.Done()
		log.Fatalln(err)
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)
	// log.Printf("img %s finished", path)
	done.Done()
}

func resizeImgNew(path string, done *sync.WaitGroup) {
	// Open the test image.
	src, err := imaging.Open(path)
	if err != nil {
		done.Done()
		log.Fatalf("Open failed: %v", err)
	}
	src = imaging.Resize(src, 3072, 0, imaging.Lanczos)
	// Save the resulting image using JPEG format.
	err = imaging.Save(src, strings.Replace(path, "resize", "new", 1))
	if err != nil {
		log.Fatalf("Save failed: %v", err)
	}
	// log.Printf("img %s finished", path)
	done.Done()
}

func resizeImgOld(path string) {
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

package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	filepath string
	language string
	dryRun   bool
)

func init() {
	flag.StringVar(&filepath, "p", "", "Path to the movie file")
	flag.StringVar(&language, "l", "pt", "Subtitle language")
	flag.BoolVar(&dryRun, "d", false, "Dry run, just try find sub")
}

/*

   function get_hash takes video path as input and returns
   hash of the video file calculated by taking first and last
   64kb of video file.


*/

func GetHash(name string) string {
	readsize := 64 * 1024
	// open file
	f, err := os.Open(name)
	if err != nil {
		fmt.Println("error")
	}
	fi, err := f.Stat()
	if err != nil {
		fmt.Println("error")
	}
	size := fi.Size()
	buf := make([]byte, readsize)
	buf1 := make([]byte, readsize)
	for {
		// read a chunk
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("error")
		}
		if n == 0 {
			break
		}
		f.Seek(size-65536, 0)
		_, err = f.Read(buf1)
		buffer := append(buf, buf1...)
		hasher := md5.New()
		hasher.Write([]byte(buffer))
		//fmt.Println(hex.EncodeToString(hasher.Sum(nil)))
		return hex.EncodeToString(hasher.Sum(nil))
	}
	return " "
}

// remove the last (right-most) dotted string
// removeExtension("video.cool.mp4") => "video.cool"
func removeExtension(fname string) string {
	chunks := strings.Split(fname, ".")
	return strings.Join(chunks[:len(chunks)-1], ".")
}

// Add a new extension for given fname
// newExtension("video.cool.mp4", "srt") => "video.cool.srt"
func newExtension(fname, extension string) string {
	return removeExtension(fname) + "." + extension
}

func SubDownloader(video_path, language string, dryRun bool) {
	hash := GetHash(video_path)
	url := "http://api.thesubdb.com/?action=download&hash=" + hash + "&language=" + language + ",en"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		req.Header.Set("User-Agent", "SubDB/1.0 (SubDownloader/0.1; http://github.com/ryukinix)")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("error: " + err.Error())
			os.Exit(1)
		}
		if resp.StatusCode == 404 {
			fmt.Printf("We did not find subtitle for %q in %q language.\n", video_path, language)
			fmt.Println("Please try any other language.")
			os.Exit(0)
		}
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("error")
		}
		if dryRun {
			fmt.Printf("Subtitle for %q in language %q found.\n", video_path, language)
		} else {
			f, err := os.Create(path.Join(path.Dir(video_path), newExtension(path.Base(video_path), "srt")))
			if err != nil {
				fmt.Println(err)
			}
			defer f.Close()
			f.Write(body)
			notify(filepath)
		}
	}
}

/*
   function notify notifies once the subtitle
   is downloaded.

*/
func notify(path string) {
	command := "notify-send"
	message := "subtitle for " + path + " downloaded!"
	cmd := exec.Command(command, message)
	_, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func main() {
	flag.Parse()
	if filepath == "" {
		fmt.Println("Please provide path to movie file")
		os.Exit(1)
	} else if t, err := exists(filepath); t == false {
		fmt.Printf("File %q didn't exists. Abort.\n", filepath)
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		os.Exit(1)
	}
	if len(language) != 2 {
		fmt.Println("invalid language, Please enter any one of these [en,es,fr,it,nl,pl,pt,ro,sv,tr]")
	}
	if dryRun {
		fmt.Println("DRY RUN MODE! THIS WILL NOT SAVE ANYTHING.")
	}
	//SubDownloader(movie_path, language)
	SubDownloader(filepath, language, dryRun)
}

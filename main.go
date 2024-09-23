package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"pixabay-downloader/pkg/api"
	"pixabay-downloader/pkg/dbmanager"
	"pixabay-downloader/pkg/model"
	log "pixabay-downloader/pkg/qlogger"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	SessionId             string
	CsfrToken             string
	AccessKey             string
	downloadFolder        string
	checkFolder           string
	shouldCheckHiddenFile bool
	wg                    sync.WaitGroup
	downloadTokens        = make(chan struct{}, 5)
	pixabay               dbmanager.DbPixabay
	tagStr                string
	tags                  []string
)

func main() {
	flag.StringVar(&SessionId, "s", "", "Web Session ID.")
	flag.StringVar(&CsfrToken, "cs", "", "CSFR Token.")
	flag.StringVar(&AccessKey, "c", "", "Client Access Key.")
	flag.StringVar(&downloadFolder, "f", "", "Folder to store images.")
	flag.StringVar(&tagStr, "t", "", "Tags use ',' as seperator. Without space.")
	flag.StringVar(&checkFolder, "x", "", "checkingFolder, 'y' or 'Y' for enable checking. Omit for skip checking.")
	flag.Parse()

	if AccessKey == "" && SessionId == "" {
		log.Panicln("Missing access key or session id or csfrtoken")
	}

	if CsfrToken == "" {
		log.Panicln("Missing csfrtoken")
	}

	if downloadFolder == "" {
		log.Panicln("Missing downloadFolder.")
	}

	if checkFolder == "y" || checkFolder == "Y" {
		shouldCheckHiddenFile = true
	} else {
		shouldCheckHiddenFile = false
	}

	if tagStr != "" {
		tags = strings.Split(tagStr, ",")
	}

	dbmanager.InitWithPath("pixabay.db").CreateTable()

	t, err := dbmanager.Dbm.GetPixabayByKey()
	if err != nil {
		tag := ""
		if len(tags) > 0 {
			tag = tags[0]
		}
		// Start from offset 0.
		t := &dbmanager.DbPixabay{
			PRKey:      "DbPixabayKey",
			CurrentTag: tag,
			PageOffset: 1,
		}
		dbmanager.Dbm.AddPixabayRecord(t)
		pixabay = *t
	} else {
		pixabay = t
	}

	if AccessKey != "" {
		for {
			downloaded := make(chan error)
			current := time.Now()

			photos, err := api.GetApiPhotos(AccessKey, int(pixabay.PageOffset))
			if err != nil {
				log.Println("GetPhotos error: ", err)
				time.Sleep(10 * time.Second)
				continue
			}

			for _, photo := range photos {
				wg.Add(1)
				log.Println("id: ", photo.ID, "url: ", photo.LargeImageURL)

				go func(p model.Hit) {
					defer wg.Done()
					error := DownloadFile(p.LargeImageURL, strconv.Itoa(int(p.ID)))
					downloaded <- error
				}(photo)
			}

			go func() {
				log.Println("Will close ch.")
				wg.Wait()
				close(downloaded)
			}()

			hasErr := false
			for anyErr := range downloaded {
				if anyErr != nil {
					hasErr = true
				}
			}
			// Next page query only for all downloads success.
			if !hasErr {
				pixabay.PageOffset++
				dbmanager.Dbm.AddPixabayRecord(&pixabay)
			}

			elapse := int32(time.Since(current).Seconds())

			log.Println("Time used: ", elapse, " seconds.")

			time.Sleep(20 * time.Second)
		}
	}

	cookies := settingCookies()

	if SessionId != "" {
		for {
			downloaded := make(chan error)
			current := time.Now()

			photos, totalPgs, err := api.CrawlerSearchPhoto(pixabay.CurrentTag, cookies, int(pixabay.PageOffset))
			if err != nil {
				log.Println("GetPhotos error: ", err)
				nextCountdown := 20
				for cnt := range nextCountdown {
					log.Println("Next try countdown: ", cnt+1)
					time.Sleep(1 * time.Second)
				}
				continue
			}

			for _, photo := range photos {
				wg.Add(1)
				// log.Println("name: ", photo)

				go func(photo string) {
					defer wg.Done()

					error := api.GetCrawlerImage(photo, downloadFolder, cookies, downloadTokens, shouldCheckHiddenFile)
					if error == nil {
						log.Println("Downloaded: ", photo)
					}
					downloaded <- error
				}(photo)
			}

			go func() {
				log.Println("Will close ch.")
				wg.Wait()
				close(downloaded)
			}()

			hasErr := false
			for anyErr := range downloaded {
				if anyErr != nil {
					hasErr = true
				}
			}

			elapse := int32(time.Since(current).Seconds())
			log.Println("Time used: ", elapse, " seconds.")

			// Next page query only for all downloads success.
			if !hasErr {
				if pixabay.PageOffset <= uint64(totalPgs) {
					pixabay.PageOffset++
				} else {
					// Crawled all photos of current tag.
					if len(tags) == 0 {
						// Start with latest of all photos.
						pixabay.PageOffset = 1
						pixabay.CurrentTag = ""
						dbmanager.Dbm.AddPixabayRecord(&pixabay)
						log.Panic("all done")

					} else {
						// Try to find next tag
						position := 0
						for idx, value := range tags {
							if value == pixabay.CurrentTag {
								position = idx
							}
						}

						// already last tag
						if position == len(tags)-1 {
							pixabay.PageOffset = 1
							pixabay.CurrentTag = ""
							dbmanager.Dbm.AddPixabayRecord(&pixabay)
							log.Panic("all done")

						}

						if position != 0 {
							// Continue with next tag
							pixabay.PageOffset = 1
							pixabay.CurrentTag = tags[position+1]
						} else {
							// Start over with new tags.
							pixabay.PageOffset = 1
							pixabay.CurrentTag = tags[0]
						}

					}
				}
				dbmanager.Dbm.AddPixabayRecord(&pixabay)
			}

			nextCountdown := 10
			fmt.Printf("Next countdown: ")
			for cnt := range nextCountdown {
				fmt.Printf("%d ", cnt+1)
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func DownloadFile(URL, fileName string) error {
	if shouldCheckHiddenFile {
		_, err := os.Stat(downloadFolder + "/.qnap")
		if err != nil {
			return err
		}
	}

	name := fileName + ".jpeg"
	_, err := os.Stat(downloadFolder + "/" + name)

	if err == nil {
		// File exist
		log.Println("Photo ", name, " already downloaded.")
		return nil
	}

	downloadTokens <- struct{}{}
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}
	// Get the response bytes from the url
	response, err := client.Get(URL)

	<-downloadTokens

	if err != nil {
		log.Println(fileName, " request error: ", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Println(fileName, " Received non 200 response code: ",
			response.StatusCode)
		return errors.New("Received non 200 response code")
	}

	log.Println("Downloaded: ", fileName)

	if shouldCheckHiddenFile {
		_, err := os.Stat(downloadFolder + "/.qnap")
		if err != nil {
			return err
		}
	}

	// Create a empty file
	// name := strconv.Itoa(int(time.Now().Unix())) + "_" + fileName + ".jpeg"
	file, err := os.Create(downloadFolder + "/" + name)
	// file, err := os.Create("./" + name)
	if err != nil {
		log.Println("Fail create file ", fileName)
		return err
	}
	defer file.Close()

	// Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Println("Fail write file ", fileName)
		return err
	}

	return nil
}

func settingCookies() []http.Cookie {
	var cookies []http.Cookie
	cookies = append(cookies, http.Cookie{
		Name:    "sessionid",
		Value:   SessionId,
		Expires: time.Now().Add(24 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "hide_ai_generated",
		Value:   "1",
		Expires: time.Now().Add(24 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "is_human",
		Value:   "1",
		Expires: time.Now().Add(24 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "g_rated",
		Value:   "off",
		Expires: time.Now().Add(24 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "csfrtoken",
		Value:   CsfrToken,
		Expires: time.Now().Add(24 * time.Hour),
	})

	return cookies
}

package main

import (
	"net/http"
	"pixabay-downloader/pkg/api"
	"pixabay-downloader/pkg/dbmanager"
	log "pixabay-downloader/pkg/qlogger"
	"testing"
	"time"
)

func TestCrawlWithTags(t *testing.T) {
	tags := []string{
		"peoples",
		"man",
		"woman",
		"boy",
		"girl",
	}

	dbmanager.InitWithPath("test.db").CreateTable()

	var cookies []http.Cookie
	cookies = append(cookies, http.Cookie{
		Name:    "sessionid",
		Value:   "",
		Expires: time.Now().Add(1 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "hide_ai_generated",
		Value:   "1",
		Expires: time.Now().Add(1 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "is_human",
		Value:   "1",
		Expires: time.Now().Add(1 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "g_rated",
		Value:   "off",
		Expires: time.Now().Add(1 * time.Hour),
	})
	cookies = append(cookies, http.Cookie{
		Name:    "csfrtoken",
		Value:   "",
		Expires: time.Now().Add(1 * time.Hour),
	})

	ttag := "dog"
	_, totalPgs, err := api.CrawlerSearchPhoto(ttag, cookies, 1)

	pixabay := &dbmanager.DbPixabay{
		PRKey:      "DbPixabayKey",
		CurrentTag: ttag,
		PageOffset: uint64(totalPgs),
	}
	dbmanager.Dbm.AddPixabayRecord(pixabay)

	if pixabay.PageOffset < uint64(totalPgs) {
		pixabay.PageOffset++
	} else {
		log.Printf("Photos for %s has been downloaded.\n", pixabay.CurrentTag)
		// Crawled all photos of current tag.
		if len(tags) == 0 {
			// Start with latest of all photos.
			dbmanager.Dbm.DeletePixabay(pixabay)
			time.Sleep(500 * time.Millisecond)
			log.Panic("No more tags to be crawled.")

		} else {
			// Try to find next tag
			log.Println("current: ", pixabay.CurrentTag, ", tags: ", tags)
			position := 0
			isCurrentTagInList := false
			for idx, value := range tags {
				if value == pixabay.CurrentTag {
					position = idx
					isCurrentTagInList = true
				}
			}

			if !isCurrentTagInList {
				// Start over with new tags.
				pixabay.PageOffset = 1
				pixabay.CurrentTag = tags[0]
				log.Printf("Remain tag has been crawled, start new tag with %s\n", pixabay.CurrentTag)

			} else {
				if position == len(tags)-1 {
					dbmanager.Dbm.DeletePixabay(pixabay)
					time.Sleep(500 * time.Millisecond)
					log.Panic("All tags has been crawled.")
				}

				// Continue with next tag
				pixabay.PageOffset = 1
				pixabay.CurrentTag = tags[position+1]
				log.Printf("Next tag: %s\n", pixabay.CurrentTag)
			}
		}
	}

	dbmanager.DeleteWithPath("test.db")

	if err != nil {
		t.Error(err)
	}
}

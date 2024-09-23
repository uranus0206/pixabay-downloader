package api

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"pixabay-downloader/pkg/model"
	log "pixabay-downloader/pkg/qlogger"
	"strconv"
)

func GetApiPhotos(key string, page int) (model.Photos, error) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	getApi := "https://pixabay.com/api"

	getApi += "?"
	getApi += "key=" + key
	getApi += "&per_page=200"
	getApi += "&page=" + strconv.Itoa(page)
	getApi += "&order=latest"
	getApi += "&image_type=photo"

	req, _ := http.NewRequest(http.MethodGet, getApi, nil)

	// q := req.URL.Query()
	// q.Add("key", key)
	// q.Add("per_page", "200")
	// q.Add("page", strconv.Itoa(page))
	// q.Add("order", "latest")
	// q.Add("image_type", "photo")
	// req.URL.RawQuery = q.Encode()

	log.Println(req.URL.String(), " , query: ", req.URL.Query())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	log.Println("Response: ", res)

	if res.StatusCode == 200 {
		// Parse links
		body, _ := io.ReadAll(res.Body)
		// log.Printf("%#v", string(body))
		pres, err := model.UnmarshalPixabaySearchResponse(body)
		log.Println("photos: ", len(pres.Hits), "err: ", err)

		if err != nil {
			return nil, err
		}

		return pres.Hits, nil
	} else {
		err = errors.New(res.Status)
		log.Println("Err: ", err)
		return nil, err
	}
}

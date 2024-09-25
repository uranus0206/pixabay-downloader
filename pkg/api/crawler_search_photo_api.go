package api

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"os"
	"pixabay-downloader/pkg/model"
	log "pixabay-downloader/pkg/qlogger"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var backoffScedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	10 * time.Second,
}

func CrawlerSearchPhoto(tag string,
	cookies []http.Cookie,
	page int,
) ([]string, int64, error) {
	var err error
	var res *http.Response

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				// accordin to https://stackoverflow.com/questions/64272533/get-request-returns-403-status-code-parsing
				// need to specify tls version, or you will get an error of 403
				MaxVersion: tls.VersionTLS12,
			},
		},
		Timeout: 30 * time.Second,
	}

	getApi := "https://pixabay.com/photos/search"

	if tag != "" {
		getApi += "/"
		getApi += tag
		getApi += "/"
	}

	getApi += "?"
	getApi += "order=latest"
	if page == 0 {
		page = 1
	}
	getApi += "&pagi=" + strconv.Itoa(page)

	req, _ := http.NewRequest(http.MethodGet, getApi, nil)

	if len(cookies) > 0 {
		for _, cookie := range cookies {
			req.AddCookie(&cookie)
		}
	}

	// req.Header.Add("Accept", "*/*")
	// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	// req.Header.Add("Cache-Control", "no-cache")
	// req.Header.Add("Connection", "keep-alive")

	log.Println(req.URL.String(), " , header: ", req.Header)

	for _, b := range backoffScedule {
		res, err = client.Do(req)

		if err == nil && res.StatusCode == 200 {
			break
		}

		log.Errorln("Request error: ", err, ", status code: ", res.StatusCode)
		log.Println("Retry search in ", b)
		time.Sleep(b)
	}

	defer res.Body.Close()

	if err != nil {
		return nil, 0, err
	}

	if res.StatusCode == 200 {
		// Parse links
		body, _ := io.ReadAll(res.Body)
		// log.Printf("%#v", string(body))
		//
		// if err != nil {
		// 	return nil, err
		// }
		//
		list, result, err := ParseSearchPhotoHtml(bytes.NewReader(body))
		log.Println("List: ", list, ", len: ", len(list), " total pages: ", result.Page.Pages, ", err: ", err)

		if err != nil {
			return nil, 0, err
		}

		if len(list) == 0 {
			return nil, 0, errors.New("no images found")
		}
		return list, result.Page.Pages, nil
	} else {
		err = errors.New(res.Status)
		log.Println("Err: ", err)
		return nil, 0, err
	}
}

func GetCrawlerImage(filename, downloadFolder string,
	cookies []http.Cookie,
	downloadToken chan struct{},
	shouldCheckHiddenFile bool,
) error {
	var err error
	var res *http.Response

	if shouldCheckHiddenFile {
		_, err := os.Stat(downloadFolder + "/.qnap")
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(downloadFolder + "/" + filename)

	if err == nil {
		// File exist
		log.Println("Photo ", filename, " already downloaded.")
		return nil
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				// accordin to https://stackoverflow.com/questions/64272533/get-request-returns-403-status-code-parsing
				// need to specify tls version, or you will get an error of 403
				MaxVersion: tls.VersionTLS12,
			},
		},
		Timeout: 30 * time.Second,
	}

	getApi := "https://pixabay.com/images/download/" + filename

	req, _ := http.NewRequest(http.MethodGet, getApi, nil)
	// req.Header.Add("Accept", "*/*")
	// req.Header.Add("Accept-Encoding", "gzip, deflate, br")

	if len(cookies) > 0 {
		for _, cookie := range cookies {
			req.AddCookie(&cookie)
		}
	}

	// for _, c := range req.Cookies() {
	// 	log.Println("Key: ", c.Name, ", Value: ", c.Value)
	// }
	//
	req.Header.Add("User-Agent", "Mozilla/5.1 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Add("Cache-Control", "no-cache")

	// Wait for semaphore
	downloadToken <- struct{}{}

	for _, b := range backoffScedule {
		res, err = client.Do(req)

		if err == nil && res.StatusCode == 200 {
			break
		}

		log.Errorln("Request error: ", err, ", status code: ", res.StatusCode)
		log.Println("Retry download ", filename, " in ", b)
		time.Sleep(b)
	}

	defer res.Body.Close()

	if err != nil {
		return err
	}

	// release semaphore
	<-downloadToken

	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode == 404 || res.StatusCode == 400 {
			log.Println(filename+" download error: ", err)
			log.Println(filename + " skip download since file or link not exist.")
			return nil
		} else {
			err = errors.New(res.Status)
			log.Println(filename+" download error: ", err)
			return err
		}
	}

	body, _ := io.ReadAll(res.Body)
	// log.Printf("%#v", string(body))

	isHtml := isHtmlDoc(bytes.NewReader(body))
	if isHtml {
		return errors.New("image not found")
	}

	if shouldCheckHiddenFile {
		_, err := os.Stat(downloadFolder + "/.qnap")
		if err != nil {
			return err
		}
	}

	// Create a empty file
	file, err := os.Create(downloadFolder + "/" + filename)
	// file, err := os.Create("./" + name)
	if err != nil {
		log.Println("Fail create file ", filename)
		return err
	}
	defer file.Close()

	// Write the bytes to the file
	written, err := io.Copy(file, bytes.NewReader(body))
	if err != nil {
		log.Println("Fail write file ", filename, ", byte: ", written)
		return err
	}
	// log.Println("write to dst: ", written)

	return nil
}

func isHtmlDoc(body io.Reader) bool {
	doc, err := html.Parse(body)
	if err != nil {
		// Cannot parse data
		log.Println("parse html error")
		return false
	}

	if doc.Type == html.DoctypeNode {
		log.Println("it's a html doc")
		return true
	}

	return false
}

func ParseSearchPhotoHtml(body io.Reader) ([]string, *model.PixabaySearchPhotoResponse, error) {
	var fileList []string
	var result *model.PixabaySearchPhotoResponse

	doc, err := html.Parse(body)
	if err != nil {
		return nil, result, err
	}

	var processAllProduct func(n *html.Node)

	processAllProduct = func(n *html.Node) {
		if n.Type == html.TextNode && strings.Contains(n.Data, "window.__BOOTSTRAP__") {
			// processNode(n)
			// log.Println("Node type: ", n.Type, ", Data: ", n.Data)
			strArray := strings.Split(n.Data, "\n")

			bootstrapStr := ""
			for _, str := range strArray {
				if strings.Contains(str, "window.__BOOTSTRAP__") {
					str = strings.ReplaceAll(str, ";", "")
					// log.Printf("str: %#v", str)
					strArr2 := strings.Split(str, ".__BOOTSTRAP__ = ")
					if len(strArr2) > 1 {
						bootstrapStr = strArr2[1]
					}
					break
				}
			}

			bootstrapStr = strings.ReplaceAll(bootstrapStr, "\"+\"", "")
			// log.Printf("bootstrap: %#v", bootstrapStr)

			resp, err := model.UnmarshalPixabaySearchPhotoResponse([]byte(bootstrapStr))
			result = &resp
			if err != nil {
				log.Printf("unmarshal error: %s", err)
				return
			}

			log.Printf("Tag: %s, Total: %d pgs, Current: %d pgs", resp.Page.Query, resp.Page.Pages, resp.Page.Page)

			for _, result := range resp.Page.Results {

				link := result.Sources.The2X

				// log.Println("link: ", link)

				linkSeps := strings.Split(link, "/")

				fileName := linkSeps[len(linkSeps)-1]
				fileName = strings.ReplaceAll(fileName, "_1280", "")
				fileList = append(fileList, fileName)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processAllProduct(c)
		}
	}

	processAllProduct(doc)

	return fileList, result, err
}

package model

/*
* https://pixabay.com/photos/search/?order=latest&pagi=1
* https://pixabay.com/images/search/people/
* {
  "request": {
    "location": {
      "canonicalHref": "https://pixabay.com/images/search/people/",
      "host": "https://pixabay.com",
      "protocol": "https",
      "pathname": "/images/search/people/",
      "path": "/images/search/people/",
      "href": "https://pixabay.com/images/search/people/"
    },
    "seo": {
      "relatedKeywords": [
        [
          "man",
          "/images/search/man/"
        ]
      ]
    }
  },
  "page": {
    "pageType": "media_list",
    "mediaType": "photo",
    "searchType": "images",
    "query": "people",
    "queryType": null,
    "suggest": "",
    "trackingQuery": "",
    "results": [
      {
        "id": 8373618,
        "width": 3375,
        "height": 6000,
        "mediaType": "photo",
        "mediaSubType": 1,
        "mediaDescriptiveType": "photo",
        "sources": {
          "1x": "https://cdn.pixabay.com/photo/2023/11/07/22/59/building-8373618_640.jpg",
          "2x": "https://cdn.pixabay.com/photo/2023/11/07/22/59/building-8373618_1280.jpg"
        }
      }
    ],
    "page": 1,
    "pages": 748,
    "total": 74721,
    "perPage": 100
  }
}
*
* */

import "encoding/json"

func UnmarshalPixabaySearchPhotoResponse(data []byte) (PixabaySearchPhotoResponse, error) {
	var r PixabaySearchPhotoResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PixabaySearchPhotoResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type PixabaySearchPhotoResponse struct {
	Request PixabaySearchPhotoResponseRequest `json:"request"`
	Page    PixabaySearchPhotoResponsePage    `json:"page"`
}

type PixabaySearchPhotoResponsePage struct {
	PageType      string                             `json:"pageType"`
	MediaType     string                             `json:"mediaType"`
	SearchType    string                             `json:"searchType"`
	Query         string                             `json:"query"`
	QueryType     interface{}                        `json:"queryType"`
	Suggest       string                             `json:"suggest"`
	TrackingQuery string                             `json:"trackingQuery"`
	Results       []PixabaySearchPhotoResponseResult `json:"results"`
	Page          int64                              `json:"page"`
	Pages         int64                              `json:"pages"`
	Total         int64                              `json:"total"`
	PerPage       int64                              `json:"perPage"`
}

type PixabaySearchPhotoResponseResult struct {
	ID                   int64                             `json:"id"`
	Width                int64                             `json:"width"`
	Height               int64                             `json:"height"`
	MediaType            string                            `json:"mediaType"`
	MediaSubType         int64                             `json:"mediaSubType"`
	MediaDescriptiveType string                            `json:"mediaDescriptiveType"`
	Sources              PixabaySearchPhotoResponseSources `json:"sources"`
}

type PixabaySearchPhotoResponseSources struct {
	The1X string `json:"1x"`
	The2X string `json:"2x"`
}

type PixabaySearchPhotoResponseRequest struct {
	Location PixabaySearchPhotoResponseLocation `json:"location"`
	SEO      PixabaySearchPhotoResponseSEO      `json:"seo"`
}

type PixabaySearchPhotoResponseLocation struct {
	CanonicalHref string `json:"canonicalHref"`
	Host          string `json:"host"`
	Protocol      string `json:"protocol"`
	Pathname      string `json:"pathname"`
	Path          string `json:"path"`
	Href          string `json:"href"`
}

type PixabaySearchPhotoResponseSEO struct {
	RelatedKeywords [][]string `json:"relatedKeywords"`
}

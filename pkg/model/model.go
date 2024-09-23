package model

import "encoding/json"

type Photos []Hit

func UnmarshalPhotos(data []byte) (Photos, error) {
	var r Photos
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Photos) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPixabaySearchResponse(data []byte) (PixabaySearchResponse, error) {
	var r PixabaySearchResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PixabaySearchResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type PixabaySearchResponse struct {
	Total     int64 `json:"total"`
	TotalHits int64 `json:"totalHits"`
	Hits      []Hit `json:"hits"`
}

type Hit struct {
	ID              int64  `json:"id"`
	PageURL         string `json:"pageURL"`
	Type            string `json:"type"`
	Tags            string `json:"tags"`
	PreviewURL      string `json:"previewURL"`
	PreviewWidth    int64  `json:"previewWidth"`
	PreviewHeight   int64  `json:"previewHeight"`
	WebformatURL    string `json:"webformatURL"`
	WebformatWidth  int64  `json:"webformatWidth"`
	WebformatHeight int64  `json:"webformatHeight"`
	LargeImageURL   string `json:"largeImageURL"`
	ImageWidth      int64  `json:"imageWidth"`
	ImageHeight     int64  `json:"imageHeight"`
	ImageSize       int64  `json:"imageSize"`
	Views           int64  `json:"views"`
	Downloads       int64  `json:"downloads"`
	Collections     int64  `json:"collections"`
	Likes           int64  `json:"likes"`
	Comments        int64  `json:"comments"`
	UserID          int64  `json:"user_id"`
	User            string `json:"user"`
	UserImageURL    string `json:"userImageURL"`
}

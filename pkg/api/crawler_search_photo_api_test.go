package api

import (
	"net/http"
	"testing"
	"time"
)

func TestCrawlerSearchPhoto(t *testing.T) {
	var cookies []http.Cookie
	cookies = append(cookies, http.Cookie{
		Name:    "sessionid",
		Value:   ".eJxVjMsOwiAQRf-FtWmgw9OfITAMKdqHKbAy_ru1C6OruzjnnifzobfJ90q7L4ldmdScW-Usu_yiGPBO64c_9u1G2IbeylwH7LVtyykO5VTXsJDfdk9LKPP39xebQp2OEgIEZWQSI-gooskiE9oxWe00VyOAkqS5VE5rl48NRiQOiIZn68BF9noDfgo95Q:1sq4pq:T6QLC_J1hs3rLs6RLKZDBDbXBBgRZPoCsPr1QjMQwQw",
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
		Value:   "RCuIUpJEhOSHw6j49np3mVNHpMjdJ9ll",
		Expires: time.Now().Add(1 * time.Hour),
	})
	_, _, err := CrawlerSearchPhoto("people", cookies, 4)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCrawlerImage(t *testing.T) {
	filename := "people-8921332.jpg"
	downloadToken := make(chan struct{}, 1)
	// err := GetCrawlerImage(filename, "", ".", downloadToken, false)

	var cookies []http.Cookie
	cookies = append(cookies, http.Cookie{
		Name:    "sessionid",
		Value:   ".eJxVjMsOwiAQRf-FtWmgw9OfITAMKdqHKbAy_ru1C6OruzjnnifzobfJ90q7L4ldmdScW-Usu_yiGPBO64c_9u1G2IbeylwH7LVtyykO5VTXsJDfdk9LKPP39xebQp2OEgIEZWQSI-gooskiE9oxWe00VyOAkqS5VE5rl48NRiQOiIZn68BF9noDfgo95Q:1sq4pq:T6QLC_J1hs3rLs6RLKZDBDbXBBgRZPoCsPr1QjMQwQw",
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
		Name:    "csfrtoken",
		Value:   "RCuIUpJEhOSHw6j49np3mVNHpMjdJ9ll",
		Expires: time.Now().Add(1 * time.Hour),
	})

	err := GetCrawlerImage(filename, ".", cookies, downloadToken, false)
	if err != nil {
		t.Error(err)
	}
}

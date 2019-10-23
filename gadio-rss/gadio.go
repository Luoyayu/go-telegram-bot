package gadio_rss

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Radios struct {
	Data     *[]RadioDataEntity `json:"data"`
	Included *[]RadioDataEntity `json:"included"`
}

type RadioDataEntity struct {
	ID            string              `json:"id"`
	Type          string              `json:"type"`
	Attributes    *RadioAttributes    `json:"attributes"`
	Relationships *RadioRelationships `json:"relationships"`
}

type RadioAttributes struct {
	// radios
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Cover       string `json:"cover"`
	PublishedAt string `json:"published-at"`
	Duration    int    `json:"duration"`

	// users
	Nickname string `json:"nickname"`
	Thumb    string `json:"thumb"`

	// categories
	Name string `json:"name"`
	//Desc        string `json:"desc"`
	Logo       string `json:"logo"`
	Background string `json:"background"`
}

type RadioRelationships struct {
	Category *RadioCategory `json:"category"`
	Djs      *RadioDJs      `json:"djs"`
}

type RadioCategory struct {
	Data *RadioDataEntity `json:"data"`
}

type RadioDJs struct {
	Data *[]RadioDataEntity `json:"data"`
}

func GetGRadios(radiosNum int) (*Radios, error) {

	params := url.Values{}
	params.Add("page[limit]", strconv.Itoa(radiosNum))
	params.Add("page[offset]", "0")
	params.Add("sort", "-published-at")
	//params.Add("include", "category,djs")
	params.Add("fields[radios]", "title,desc,published-at,duration")
	//params.Add("fields[radios]", "title,desc,cover,published-at,duration,category,djs")

	radioUrl, _ := url.Parse("http://www.gcores.com/gapi/v1/radios?" + params.Encode())
	log.Println(radioUrl)

	resp, err := http.Get(radioUrl.String())

	if err != nil {
		log.Println(err)
	} else {
		var resultMap Radios
		d := json.NewDecoder(resp.Body)
		err := d.Decode(&resultMap)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("decode body of response ok!")
			return &resultMap, nil
		}
	}
	return &Radios{}, err
}

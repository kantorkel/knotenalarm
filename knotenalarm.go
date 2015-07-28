package main

import (
	"encoding/json"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/scalingdata/gcfg"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	Anaconda struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}
	Freifunk struct {
		NodelistUrl string
		MapUrl      string
	}
	Config struct {
		Email          string
		UpdateInterval time.Duration
		Debug          bool
	}
}

type nodelist struct {
	Nodes []struct {
		Status struct {
			Lastcontact  string `json:"lastcontact"`
			Clients      int    `json:"clients"`
			Firstcontact string `json:"firstcontact"`
			Online       bool   `json:"online"`
		} `json:"status"`
		Position struct {
			Lat  float64 `json:"lat"`
			Long float64 `json:"long"`
		} `json:"position"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"nodes"`
	Version   string `json:"version"`
	UpdatedAt string `json:"updated_at"`
}

type address struct {
	PlaceID     string `json:"place_id"`
	Licence     string `json:"licence"`
	OsmType     string `json:"osm_type"`
	OsmID       string `json:"osm_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
	Address     struct {
		Marina       string `json:"marina"`
		HouseNumber  string `json:"house_number"`
		Road         string `json:"road"`
		Suburb       string `json:"suburb"`
		CityDistrict string `json:"city_district"`
		City         string `json:"city"`
		Town         string `json:"town"`
		Village      string `json:"village"`
		State        string `json:"state"`
		Postcode     string `json:"postcode"`
		Country      string `json:"country"`
		CountryCode  string `json:"country_code"`
	} `json:"address"`
}

func LoadJson(URL string, v interface{}) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("LoadJson: response not OK: %s", response.Status)
	}

	return json.NewDecoder(response.Body).Decode(&v)
}

func main() {
	var cfg Config
	var locationName string
	var nodeList nodelist
	var nodeAddress address

	check(gcfg.ReadFileInto(&cfg, "myconfig.gcfg"))

	now := time.Now().Add(-cfg.Config.UpdateInterval * time.Hour)

	anaconda.SetConsumerKey(cfg.Anaconda.ConsumerKey)
	anaconda.SetConsumerSecret(cfg.Anaconda.ConsumerSecret)
	api := anaconda.NewTwitterApi(cfg.Anaconda.AccessToken, cfg.Anaconda.AccessTokenSecret)

	check(LoadJson(cfg.Freifunk.NodelistUrl, &nodeList))

	for i := range nodeList.Nodes {
		t, err := time.Parse("2006-01-02T15:04:05", nodeList.Nodes[i].Status.Firstcontact)
		check(err)

		if t.After(now) && nodeList.Nodes[i].Position.Lat != 0 {
			jsonUrl := "https://nominatim.openstreetmap.org/reverse?format=json&lat=" + strconv.FormatFloat(nodeList.Nodes[i].Position.Lat, 'f', 5, 64) + "&lon=" + strconv.FormatFloat(nodeList.Nodes[i].Position.Long, 'f', 5, 64) + "&zoom=16&addressdetails=1"

			if len(cfg.Config.Email) > 0 {
				jsonUrl += "&email=" + cfg.Config.Email
			}
			check(LoadJson(jsonUrl, &nodeAddress))

			if nodeAddress.Address.Suburb != "" {
				locationName = nodeAddress.Address.Suburb
			} else if nodeAddress.Address.CityDistrict != "" {
				locationName = nodeAddress.Address.CityDistrict
			} else if nodeAddress.Address.Village != "" {
				locationName = nodeAddress.Address.Village
			} else if nodeAddress.Address.Town != "" {
				locationName = nodeAddress.Address.Town
			} else if nodeAddress.Address.City != "" {
				locationName = nodeAddress.Address.City
			} else {
				locationName = "*hust*"
			}

			if cfg.Config.Debug {
				fmt.Println(locationName + " " + nodeList.Nodes[i].Name + " " + cfg.Freifunk.MapUrl + "#!v:m;n:" + nodeList.Nodes[i].ID)
			} else {
				update, err := api.PostTweet("In "+locationName+" gibt es einen neuen #Freifunk-Knoten: "+nodeList.Nodes[i].Name+" "+cfg.Freifunk.MapUrl+"#!v:m;n:"+nodeList.Nodes[i].ID, nil)
				check(err)
				fmt.Println(update)
			}
			time.Sleep(1100 * time.Millisecond)
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

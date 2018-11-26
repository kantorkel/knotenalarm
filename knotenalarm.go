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
	Version int `json:"version"`
	Nodes   []struct {
		Nodeinfo struct {
			System struct {
				SiteCode string `json:"site_code"`
			} `json:"system"`
			Vpn      bool   `json:"vpn"`
			Hostname string `json:"hostname"`
			Software struct {
				BatmanAdv struct {
					Version string `json:"version"`
				} `json:"batman-adv"`
				Fastd struct {
					Enabled bool   `json:"enabled"`
					Version string `json:"version"`
				} `json:"fastd"`
				Firmware struct {
					Release string `json:"release"`
					Base    string `json:"base"`
				} `json:"firmware"`
			} `json:"software"`
			NodeID   string `json:"node_id"`
			Hardware struct {
				Nproc int `json:"nproc"`
			} `json:"hardware"`
			Location struct {
				Latitude float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			Network struct {
				Mac       string   `json:"mac"`
				Addresses []string `json:"addresses"`
				Mesh      struct {
					BatHhwest struct {
						Interfaces struct {
							Tunnel []string `json:"tunnel"`
						} `json:"interfaces"`
					} `json:"bat-hhwest"`
				} `json:"mesh"`
			} `json:"network"`
		} `json:"nodeinfo"`
		Flags struct {
			Online bool `json:"online"`
		} `json:"flags"`
		Statistics struct {
			Uptime      float64 `json:"uptime"`
			MemoryUsage float64 `json:"memory_usage"`
			Clients     int     `json:"clients"`
			Loadavg     float64 `json:"loadavg"`
		} `json:"statistics"`
		Lastseen  time.Time `json:"lastseen"`
		Firstseen time.Time `json:"firstseen"`
	} `json:"nodes"`
	Timestamp time.Time `json:"timestamp"`
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

	check(gcfg.ReadFileInto(&cfg, "myconfig.gcfg"))

	now := time.Now().Add(-cfg.Config.UpdateInterval * time.Minute)

	anaconda.SetConsumerKey(cfg.Anaconda.ConsumerKey)
	anaconda.SetConsumerSecret(cfg.Anaconda.ConsumerSecret)
	api := anaconda.NewTwitterApi(cfg.Anaconda.AccessToken, cfg.Anaconda.AccessTokenSecret)

	check(LoadJson(cfg.Freifunk.NodelistUrl, &nodeList))

	for i := range nodeList.Nodes {
		t := nodeList.Nodes[i].Firstseen

		if t.After(now) && nodeList.Nodes[i].Nodeinfo.Location.Latitude != 0 {
			jsonUrl := "https://nominatim.openstreetmap.org/reverse?format=json&lat=" + strconv.FormatFloat(nodeList.Nodes[i].Nodeinfo.Location.Latitude, 'f', 5, 64) + "&lon=" + strconv.FormatFloat(nodeList.Nodes[i].Nodeinfo.Location.Longitude, 'f', 5, 64) + "&zoom=16&addressdetails=1"

			if len(cfg.Config.Email) > 0 {
				jsonUrl += "&email=" + cfg.Config.Email
			}

			var nodeAddress address
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
				fmt.Println(locationName + " " + nodeList.Nodes[i].Nodeinfo.Hostname + " " + cfg.Freifunk.MapUrl + "#!v:m;n:" + nodeList.Nodes[i].Nodeinfo.NodeID)
			} else {
				update, err := api.PostTweet("In "+locationName+" gibt es einen neuen #Freifunk-Knoten: "+nodeList.Nodes[i].Nodeinfo.Hostname+" "+cfg.Freifunk.MapUrl+"#!v:m;n:"+nodeList.Nodes[i].Nodeinfo.NodeID, nil)
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

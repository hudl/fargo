package eurekago

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getJson(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// At this point we're done and shit worked, simply return the bytes
	return body, nil
}

func (e *EurekaConnection) GetApps() *EurekaApps {
	url := fmt.Sprintf("%s://%s:%s/%s", e.Proto, e.Address, e.Port, e.Urls.Apps)
	println(url)
	out, err := getJson(url)
	if err != nil {
		println("Couldn't get JSON.", err.Error())
	}
	println(string(out))
	var v interface{}
	json.Unmarshal(out, &v)
	return &EurekaApps{}
}

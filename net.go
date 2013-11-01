package eurekago

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func getJson(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

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
	return &EurekaApps{}
}

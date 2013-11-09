package eurekago

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getXml(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/xml")

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

func (e *EurekaConnection) GetApps() {
	url := fmt.Sprintf("%s://%s:%s/%s", e.Proto, e.Address, e.Port, e.Urls.Apps)
	out, err := getXml(url)
	if err != nil {
		println("Couldn't get XML.", err.Error())
	}
	fmt.Println(string(out))
	var v GetAppsResponse
	err = xml.Unmarshal(out, &v)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(v)
	fmt.Println(v.Applications[0].Instances[0].LeaseInfo)
}

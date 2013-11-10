package eurekago

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

func postXml(url string, reqBody []byte) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, -1, err
	}
	body, rcode, err := reqXml(req)
	if err != nil {
		return nil, rcode, err
	}
	return body, rcode, nil
}

func getXml(url string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, -1, err
	}
	body, rcode, err := reqXml(req)
	if err != nil {
		return nil, rcode, err
	}
	return body, rcode, nil
}

func reqXml(req *http.Request) ([]byte, int, error) {

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Accept", "application/xml")

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, err
	}
	// At this point we're done and shit worked, simply return the bytes
	return body, resp.StatusCode, nil
}

func (e *EurekaConnection) GetApps() map[string]Application {
	url := fmt.Sprintf("%s://%s:%s/%s", e.Proto, e.Address, e.Port, e.Urls.Apps)
	out, rcode, err := getXml(url)
	if err != nil {
		println("Couldn't get XML.", err.Error())
	}
	fmt.Println(string(out))
	var v GetAppsResponse
	err = xml.Unmarshal(out, &v)
	if err != nil {
		fmt.Println(err.Error())
	}
	apps := map[string]Application{}
	for _, app := range v.Applications {
		apps[app.Name] = app
	}
	if rcode > 299 || rcode < 200 {
		fmt.Println("Non-200 rcode of " + string(rcode))
	}
	return apps
}

func (e *EurekaConnection) RegisterInstance(ins *Instance) error {
	url := fmt.Sprintf("%s://%s:%s/%s", e.Proto, e.Address, e.Port, e.Urls.Apps)

	_, rcode, err := getXml(url + "/" + ins.HostName)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if rcode == 200 {
		fmt.Println("Instance exists. NOOP")
	} else {
		fmt.Println("Instance not yet registered. Registering.")
	}

	out, err := xml.Marshal(ins)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, rcode, err = postXml(url, out)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if rcode != 204 {
		fmt.Println("HTTP returned " + string(rcode) + " possible failure creating instance")
	}

	body, rcode, err := getXml(url + "/" + ins.HostName)
	xml.Unmarshal(body, ins)
	return nil
}

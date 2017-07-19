package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var HttpClient = &http.Client{
	Transport: transport,
	Timeout:   30 * time.Second,
}

var transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	ResponseHeaderTimeout: 10 * time.Second,
}

func postBody(reqURL string, reqBody []byte, isJson bool) ([]byte, int, error) {
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(reqBody))
	if err != nil {
		log.Errorf("Could not create POST %s with body %s, error: %s", reqURL, string(reqBody), err.Error())
		return nil, -1, err
	}
	log.Debugf("postBody: %s %s : %s\n", req.Method, req.URL, string(reqBody))
	body, rcode, err := netReqTyped(req, isJson)
	if err != nil {
		log.Errorf("Could not complete POST %s with body %s, error: %s", reqURL, string(reqBody), err.Error())
		return nil, rcode, err
	}
	//eurekaCache.Flush()
	return body, rcode, nil
}

func putKV(reqURL string, pairs map[string]string) ([]byte, int, error) {
	params := url.Values{}
	for k, v := range pairs {
		params.Add(k, v)
	}
	parameterizedURL := reqURL + "?" + params.Encode()
	log.Noticef("Sending KV request with URL %s", parameterizedURL)
	req, err := http.NewRequest("PUT", parameterizedURL, nil)
	if err != nil {
		log.Errorf("Could not create PUT %s, error: %s", reqURL, err.Error())
		return nil, -1, err
	}
	body, rcode, err := netReq(req) // TODO(cq) I think this can just be netReq() since there is no body
	if err != nil {
		log.Errorf("Could not complete PUT %s, error: %s", reqURL, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func getBody(reqURL string, isJson bool) ([]byte, int, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Errorf("Could not create GET %s, error: %s", reqURL, err.Error())
		return nil, -1, err
	}
	body, rcode, err := netReqTyped(req, isJson)
	if err != nil {
		log.Errorf("Could not complete GET %s, error: %s", reqURL, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func deleteReq(reqURL string) (int, error) {
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		log.Errorf("Could not create DELETE %s, error: %s", reqURL, err.Error())
		return -1, err
	}
	_, rcode, err := netReq(req)
	if err != nil {
		log.Errorf("Could not complete DELETE %s, error: %s", reqURL, err.Error())
		return rcode, err
	}
	return rcode, nil
}

func netReqTyped(req *http.Request, isJson bool) ([]byte, int, error) {
	if isJson {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/xml")
		req.Header.Set("Accept", "application/xml")
	}
	return netReq(req)
}

func netReq(req *http.Request) ([]byte, int, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = HttpClient.Do(req)
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			// it's a transient network error so we sleep for a bit and try
			// again in case it's a short-lived issue
			log.Warningf("Retrying after temporary network failure, error: %s",
				nerr.Error())
			time.Sleep(10)
		} else {
			break
		}
	}
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failure reading request body, error: %s", err.Error())
		return nil, -1, err
	}
	// At this point we're done and shit worked, simply return the bytes
	log.Infof("Got eureka response from url=%v", req.URL)
	return body, resp.StatusCode, nil
}

package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"bytes"
	"github.com/mreiferson/go-httpclient"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

func postXML(reqURL string, reqBody []byte) ([]byte, int, error) {
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(reqBody))
	if err != nil {
		log.Error("Could not create POST %s with body %s Error: %s", reqURL, string(reqBody), err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Could not complete POST %s with body %s Error: %s", reqURL, string(reqBody), err.Error())
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
	log.Notice("Sending metadata request with URL %s", parameterizedURL)
	req, err := http.NewRequest("PUT", parameterizedURL, nil)
	if err != nil {
		log.Error("Could not create PUT %s with Error: %s", reqURL, err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Could not complete PUT %s with Error: %s", reqURL, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func getXML(reqURL string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Error("Could not create GET %s with Error: %s", reqURL, err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Could not complete GET %s with Error: %s", reqURL, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func reqXML(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Accept", "application/xml")
	return netReq(req)
}

func netReq(req *http.Request) ([]byte, int, error) {
	transport := &httpclient.Transport{
		ConnectTimeout:        5 * time.Second,
		RequestTimeout:        30 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	defer transport.Close()

	// Send the request via a client
	client := &http.Client{Transport: transport}
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			// it's a transient network error so we sleep for a bit and try
			// again in case it's a short-lived issue
			log.Warning("Retrying after temporary network failure, error: %s",
				nerr.Error())
			time.Sleep(10)
		}
	}
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failure reading request body Error: %s", err.Error())
		return nil, -1, err
	}
	// At this point we're done and shit worked, simply return the bytes
	return body, resp.StatusCode, nil
}

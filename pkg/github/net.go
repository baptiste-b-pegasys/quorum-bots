package github

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type httpAdapter struct {
	httpClient *http.Client
}

func (adapter *httpAdapter) sendPostRequest(url string, body io.Reader) ([]byte, error) {
	fmt.Println(url)
	req, _ := http.NewRequest("POST", url, body)
	return adapter.sendRequest(req)
}

func (adapter *httpAdapter) sendGetRequest(url string) ([]byte, error) {
	fmt.Println(url)
	req, _ := http.NewRequest("GET", url, nil)
	return adapter.sendRequest(req)
}

func (adapter *httpAdapter) sendRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(USERNAME, TOKEN)
	req.Header.Add("Accept", "application/vnd.v3+json")
	resp, _ := adapter.httpClient.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// print HTTP headers
	//for name, values := range resp.Header {
	//	// Loop over all values for the name.
	//	for _, value := range values {
	//		fmt.Println(name, value)
	//	}
	//}

	return body, nil
}

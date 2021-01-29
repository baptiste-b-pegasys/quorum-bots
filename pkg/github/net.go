package github

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const username = "ricardolyn"
const token = "8b3d3c6b486590135699987e7e760de92575c8bf"

type httpAdapter struct {
	httpClient *http.Client
}

func newHttpAdapter() *httpAdapter {
	client := &http.Client{}
	return &httpAdapter{
		httpClient: client,
	}
}

func (adapter *httpAdapter) sendPostRequest(url string, body io.Reader) ([]byte, error) {
	req, _ := http.NewRequest("POST", url, body)
	return adapter.sendRequest(req)
}

func (adapter *httpAdapter) sendGetRequest(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	return adapter.sendRequest(req)
}

func (adapter *httpAdapter) sendRequest(req *http.Request) ([]byte, error) {
	log.Printf("%s %s\n", req.Method, req.URL)
	req.SetBasicAuth(username, token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, _ := adapter.httpClient.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// For DEBUG. print HTTP headers
	//for name, values := range resp.Header {
	//	// Loop over all values for the name.
	//	for _, value := range values {
	//		fmt.Println(name, value)
	//	}
	//}

	return body, nil
}

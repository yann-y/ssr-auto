package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zztroot/rconfig"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type nodeList struct {
	Proxies struct {
		Proxy struct {
			All     []string      `json:"all"`
			History []interface{} `json:"history"`
			Name    string        `json:"name"`
			Now     string        `json:"now"`
			Type    string        `json:"type"`
			UDP     bool          `json:"udp"`
		} `json:"Proxy"`
	} `json:"proxies"`
}
type delay struct {
	Delay int `json:"delay,omitempty"`
}

func getProxies(host string) *nodeList {
	url := host + "/proxies"

	log.Println("请求地址:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	data, err := ioutil.ReadAll(resp.Body)
	getNodeList := nodeList{}
	err = json.Unmarshal(data, &getNodeList)
	if err != nil {
		return nil
	}
	fmt.Printf(getNodeList.Proxies.Proxy.Now)
	timeout := 5000
	wg := sync.WaitGroup{}
	requestUrl := "http://www.gstatic.com/generate_204"
	nodeMap := sync.Map{}
	ch := make(chan struct{}, 10)
	for _, value := range getNodeList.Proxies.Proxy.All {
		v := value
		wg.Add(1)
		ch <- struct{}{}
		go func(v string) {
			defer wg.Done()
			requestsUrl := url + "/" + v + "/delay?" + "timeout=" + strconv.Itoa(timeout) + "&url=" + requestUrl
			get, err := http.Get(requestsUrl)
			if err != nil {
				return
			}
			defer get.Body.Close()
			data, _ := ioutil.ReadAll(get.Body)
			//log.Println(string(data))
			temp := delay{}
			err = json.Unmarshal(data, &temp)
			if err != nil {
				return
			}
			nodeMap.Store(v, temp.Delay)
			<-ch
		}(v)
	}
	wg.Wait()
	minName := ""
	minDelay := 9999
	nodeMap.Range(func(key, value interface{}) bool {
		fmt.Println(key, value)
		if value.(int) <= minDelay && value.(int) > 0 {
			minName = key.(string)
			minDelay = value.(int)
		}
		return true
	})
	fmt.Println(minName, minDelay)
	changeNode(url, minName)
	return &getNodeList
}
func changeNode(url, nodeName string) {
	url = url + "/Proxy"
	client := &http.Client{}
	postData := fmt.Sprintf("{\"name\":\"%s\"}", nodeName)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(postData)))
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fmt.Println("status", resp.Status)
}
func main() {
	files, _ := rconfig.OpenJson("./config.json")
	name := files.GetString("host") //key
	//desc := files.GetInt("test.1.params.0.desc")  //333
	getProxies(name)
}

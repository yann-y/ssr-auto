package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
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
func GetFileCreateTime(path string) int64 {
	osType := runtime.GOOS
	fileInfo, _ := os.Stat(path)
	if osType == "linux" {
		statT := fileInfo.Sys().(*syscall.Stat_t)
		tCreate := statT.Ctim.Sec
		return tCreate
	}
	return time.Now().Unix()
}
func getClash(url string) {
	// 比较上一次的更新时间
	localPath := "./config.yaml"
	t := time.Now().Unix() - GetFileCreateTime(localPath)
	fmt.Println(t)
	if t <= 60*60*24 {
		log.Println("距离上一次更新配置文件太近")
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// 打开配置问价
	open, err := os.OpenFile("./config.yaml", os.O_CREATE|os.O_RDWR, 0766)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = io.Copy(open, resp.Body)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	fmt.Printf(strconv.FormatInt(GetFileCreateTime("./config.yaml"), 10))
}
func main() {
	url := ""
	getClash(url)
	//files, _ := rconfig.OpenJson("./config.json")
	//name := files.GetString("host") //key
	//desc := files.GetInt("test.1.params.0.desc")  //333

	//getProxies(name)
}

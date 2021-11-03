package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zztroot/rconfig"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var filePath = "/home/yanfive/Pictures/clash/config.yaml"

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
	//fmt.Println(minName, minDelay)
	log.Println("=====测速完成！！！=====")
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
	status := resp.Status
	if status == "204 No Content" {
		log.Println("更新节点成功！！")
		log.Println("", nodeName)
	}
	fmt.Println("status", status)
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
func getClash(url string, ok bool) {
	// 比较上一次的更新时间
	localPath := filePath
	t := time.Now().Unix() - GetFileCreateTime(localPath)
	if t <= 60*60*24 && ok {
		log.Println("距离上一次更新配置文件太近")
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	// 打开配置问价
	open, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0766)
	if err != nil {
		log.Println("url错误------》", err)
		return
	}

	_, err = io.Copy(open, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	//fmt.Printf(strconv.FormatInt(GetFileCreateTime("./config.yaml"), 10))
	log.Println("更新时间配置文件成功")
	restart()
	time.Sleep(30 * time.Second)
}
func restart() {
	cmd := exec.Command("systemctl", "restart", `clash`)
	//创建获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
		return
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err,", err)
		return
	}
	//使用带缓冲的读取器
	outputBuf := bufio.NewReader(stdout)
	for {
		//一次获取一行,_ 获取当前行是否被读完
		output, _, err := outputBuf.ReadLine()
		if err != nil {

			// 判断是否到文件的结尾了否则出错
			if err.Error() != "EOF" {
				fmt.Printf("Error :%s\n", err)
			}
			return
		}
		fmt.Printf("%s\n", string(output))
	}
}
func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("获取当前运行路径出现错误")
	}
	dir = path.Join(dir, "config.json")
	files, _ := rconfig.OpenJson(dir)
	url := files.GetString("clash")
	filePath = files.GetString("yaml")
	getClash(url, true)
	//files, _ := rconfig.OpenJson("./config.json")
	name := files.GetString("host") //key
	getProxies(name)
}

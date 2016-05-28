/*
* @Author: tudou
* @Date:   2016-05-26 10:11:03
* @Last Modified by:   xiwang
* @Last Modified time: 2016-05-28 17:33:23
 */

package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//HostServer host server
type HostServer struct {
	Host   string
	Server string
	Name   string
}

func newHostServer(host string, server string, name string) *HostServer {
	return &HostServer{Host: host, Server: server, Name: name}
}

func findLink(weburl string) (map[string]string, error) {
	doc, err := goquery.NewDocument(weburl)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	stores := make(map[string]string)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		text := s.Text()
		if !strings.HasPrefix(link, "http") {
			link = weburl + link
		}

		stores[link] = text
	})

	log.Printf("Find %d links ", len(stores))
	return stores, nil
}

func headRequest(weburl string, name string) {
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, _ := http.NewRequest("HEAD", weburl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		hostChannel <- HostServer{}
		return
	}

	server := resp.Header.Get("server")
	u, _ := url.Parse(weburl)
	host := u.Host
	hostServer := *newHostServer(host, server, name)
	hostChannel <- hostServer
}

var hostChannel chan HostServer

func analysisHostServer(weburl string) {
	links, err := findLink(weburl)
	if err == nil {
		for link, name := range links {
			fmt.Printf("requested %s \n", link)
			go headRequest(link, name)
		}

		datas := make(map[string]HostServer)
		total := len(links)
		i := 0
	OutLooper:
		for {
			data := <-hostChannel
			i++
			fmt.Printf("----%d of %d---received %s \n", i, total, data.Host)
			datas[data.Host] = data
			if i >= len(links) {
				break OutLooper
			}
		}

		i = 0
		for _, data := range datas {
			i++
			fmt.Printf("--%d-- %s --- %s(%s)\n", i, data.Host, data.Server, data.Name)
		}
	}
}

func main() {
	hostChannel = make(chan HostServer, 5)
	analysisHostServer("http://www.265.com/")
}

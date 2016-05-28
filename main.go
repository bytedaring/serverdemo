/*
* @Author: tudou
* @Date:   2016-05-26 10:11:03
* @Last Modified by:   tudou
* @Last Modified time: 2016-05-28 09:28:01
 */

package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	//	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	resp, err := http.Head(weburl)
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
		for key, link := range links {
			fmt.Printf("requested %s \n", link)
			go headRequest(key, link)
		}

		datas := make(map[string]HostServer)
		total := len(links)
		i := 0
	OutLooper:
		for {
			i++
			select {
			case data := <-hostChannel:
				fmt.Printf("----%d of %d---received %s \n", i, total, data.Host)
				datas[data.Host] = data
				if i >= len(links)-1 {
					break OutLooper
				}
			}
		}
		fmt.Println("---end--")
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

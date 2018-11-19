//package rest_free_proxy_list
package main

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"bytes"
	"golang.org/x/net/html"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"encoding/json"
)

const (
	// URLs
	freeProxyListURL = "https://free-proxy-list.net/"
	saveProxyBatchURL = "https://rest-free-proxy-list.herokuapp.com/proxies/batch"
	deleteProxyBatchURL = "https://rest-free-proxy-list.herokuapp.com/proxies"

	// Local storage path
	htmlPageStorage = "/tmp/free-proxy-list"
	htmlFilePath = htmlPageStorage + "/index.html"

)

type Batch struct {
	Batch 	[]Proxy  `json:"batch"`
}

type Proxy struct {
	IP          string	`json:"ip"`
	Port        int		`json:"port"`
	Code        string 	`json:"code"`
	Country     string	`json:"country"`
	Anonymity   string	`json:"anonymity"`
	Google      bool	`json:"google"`
	HTTPS       bool	`json:"https"`
	LastChecked string	`json:"last_checked"`
}

func (p Proxy) String() string {
	return fmt.Sprintf("%-15s |  %-5d | %-2s | %-20s | %-10s | %t | %t | %s\n",
		p.IP, p.Port, p.Code, p.Country, p.Anonymity, p.Google, p.HTTPS, p.LastChecked)
}

func getPage() []byte{
	u, err := url.Parse(freeProxyListURL)
	if err != nil { panic(err) }

	r, err := http.Get(u.String())
	if err != nil { panic(err) }

	// Save response body in []byte so it may be reused
	pageBytes, err := ioutil.ReadAll(r.Body)
	if err != nil { panic(err) }

	return pageBytes
}

func savePage(body []byte){
	err := os.MkdirAll(htmlPageStorage, 0777)
	if err != nil { panic(err) }

	err = ioutil.WriteFile(htmlFilePath, body, 0777)
	if err != nil { panic(err) }

	fmt.Printf("Saved at %s\n", htmlFilePath)
}

func containsId(n *html.Node, id string) bool {
	for  _, a := range n.Attr {
		if a.Key == "id" {
			return a.Val == id
		}
	}
	return false
}

func parseTableCell(tCell *html.Node) string{
	var b bytes.Buffer

	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(tCell)

	return b.String()
}

func parseTableRow(tRow *html.Node) Proxy{

	nProxy := Proxy{}
	index := 0
	for c := tRow.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			cell := parseTableCell(c)
			if index == 0 && isIP(cell) {
				nProxy.IP = cell
			} else if index == 1 && isPort(cell) {
				p, err := strconv.Atoi(cell)
				if err != nil {panic(err)}
				nProxy.Port = p
			} else if index == 2 && isCountryCode(cell) {
				nProxy.Code = cell
			} else if index == 3 && isCountryName(cell) {
				nProxy.Country = cell
			} else if index == 4 && isAnonymity(cell) {
				nProxy.Anonymity = cell
			} else if index == 5 && hasGoogle(cell) {
				if nProxy.Google = true; cell == "no" {
					nProxy.Google = false
				}
			} else if index == 6 && hasHttps(cell) {
				if nProxy.HTTPS = true; cell == "no" {
					nProxy.HTTPS = false
				}
			} else if index == 7 && isUpdateComment(cell) {
				nProxy.LastChecked = cell
			}
		}
		index++
	}

	return nProxy
}

func parseTableBody(tBody *html.Node) []Proxy{
	var proxies []Proxy
	for c := tBody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			proxies = append(proxies, parseTableRow(c))
		}
	}
	return proxies
}

func parseTable(table *html.Node) []Proxy{
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			return parseTableBody(c)
		}
	}
	return []Proxy{}
}

func findTableNode(pageBytes []byte) (tNode *html.Node){
	reader := bytes.NewReader(pageBytes)

	doc, err := html.Parse(reader)
	if err != nil { panic(err) }

	var findTable func(*html.Node) bool
	findTable = func(n *html.Node) bool{
		if n.Type == html.ElementNode && n.Data == "table" && containsId(n, "proxylisttable") {
			tNode = n
			return true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling{
			if findTable(c) {
				// This return prevents a unnecessary checks once the table has been found
				return true
			}
		}
		return false
	}
	findTable(doc) // if this is true, it found the table. TODO: Test it

	return
}

func isIP(ip string) bool {
	ipRe := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	return ipRe.MatchString(ip)
}

func isPort(port string) bool{
	portRe := regexp.MustCompile(`^\d{2,5}$`)
	return portRe.MatchString(port)
}

func isCountryCode(code string) bool {
	codeRe := regexp.MustCompile(`^[A-Z]{2}$`)
	return codeRe.MatchString(code)
}

func isCountryName(name string) bool {
	nameRe := regexp.MustCompile(`^[A-Z](\w|\s)+$`)
	return nameRe.MatchString(name)
}

func isAnonymity(str string) bool {
	// TODO: This should eval from an ENUM [transparent, anonymity, elite proxy, (find the rest) ]
	anonRe := regexp.MustCompile(`.*`)
	return anonRe.MatchString(str)
}

func hasGoogle(str string) bool {
	googRe := regexp.MustCompile(`^yes|no$`)
	return googRe.MatchString(str)
}

func hasHttps(str string) bool {
	httRe := regexp.MustCompile(`^yes|no$`)
	return httRe.MatchString(str)
}

func isUpdateComment(comment string) bool {
	// TODO:
	// Look for specific, constant words that should be in this string:
	// [ago, minutes, seconds, hours, *digits]
	commentRe := regexp.MustCompile(`.*`)
	return commentRe.MatchString(comment)
}

func deleteAllProxies() bool {
	req, err := http.NewRequest(http.MethodDelete, deleteProxyBatchURL, nil)
	if err != nil { panic(err) }
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil { panic(err) }

	fmt.Println(resp.StatusCode)
	return resp.StatusCode == http.StatusNoContent
}

func saveProxies(proxies []Proxy)  {
	batch := Batch{ Batch: proxies }
	b, err := json.Marshal(batch)
	//b, err := json.MarshalIndent(proxies, "", " ")
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, saveProxyBatchURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil { panic(err) }
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func main() {
	pageBytes := getPage()
	savePage(pageBytes)
	tableNode := findTableNode(pageBytes)
	proxies := parseTable(tableNode)

	if deleteAllProxies() {
		saveProxies(proxies)
	}else {
		fmt.Printf("Could not clean up. Halting All inserts until DELETE is fixed")
	}
}
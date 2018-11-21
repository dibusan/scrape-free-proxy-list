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
	"time"
	"log"
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

const (
	ERROR	=  iota // 0
	WARNING			// 1
	INFO			// 2
	DEBUG			// 3
	VERBOSE			// 4
)

var (
	runtimeLog *os.File
	logger *log.Logger
	logLevel = ERROR
)

func initLogger(ll int){
	logLevel = ll
	runtimeLog, err := os.OpenFile("/var/log/scrapefreeproxylist.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	logger = log.New(runtimeLog, "applog: ", log.Lshortfile|log.LstdFlags)

}

func logError(msg string){
	logger.Println("ERROR: " + msg + "\nExit Status 1")
}

func logWarning(msg string){
	if(logLevel >= WARNING){
		logger.Println("WARNING: " + msg)
	}
}

func logInfo(msg string){
	if(logLevel >= INFO){
		logger.Println("INFO: " + msg)
	}
}

func logDebug(msg string){
	if(logLevel >= DEBUG){
		logger.Println("DEBUG: " + msg)
	}
}

func logVerbose(msg string){
	if(logLevel == VERBOSE){
		logger.Println("VERBOSE: " + msg)
	}
}

func getPage() ([]byte, error){
	logInfo("Getting Free Proxy List webpage over HTTP")
	u, err := url.Parse(freeProxyListURL)
	if err != nil {
		logError("Failed to parse URL " + freeProxyListURL)
		os.Exit(1)
	}

	r, err := http.Get(u.String())
	if err != nil {
		logError("Failed HTTP Request. Err: " + err.Error())
		return nil, err
	}

	// Save response body in []byte so it may be reused
	pageBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logError("Failed to read HTTP Response. Err: " + err.Error())
		return nil, err
	}

	logVerbose("Successfully retrieved Free Proxy List page contents.")
	return pageBytes, nil
}

func savePage(body []byte){
	err := os.MkdirAll(htmlPageStorage, 0777)
	if err != nil {
		logWarning("Failed to make Html Page Directory. Err: " + err.Error())
		return
	}

	err = ioutil.WriteFile(htmlFilePath, body, 0777)
	if err != nil {
		logWarning("Failed to save HTML body. Err: " + err.Error())
		return
	}

	logVerbose("Saved HTML body at " + htmlFilePath)
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
	logInfo("Parse Row")

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
	logInfo("Extract proxies from table body")
	var proxies []Proxy
	for c := tBody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			proxies = append(proxies, parseTableRow(c))
		}
	}
	return proxies
}

func parseTable(table *html.Node) []Proxy{
	logInfo("Parse table")
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			return parseTableBody(c)
		}
	}
	logWarning("Table body not found.")
	//TODO: logNotify
	//logNotify("Page structure might have changed", pageBytes, err)

	return []Proxy{}
}

func findTableNode(pageBytes []byte) (tNode *html.Node, err error){
	logInfo("Finding proxies table node")
	reader := bytes.NewReader(pageBytes)

	doc, err := html.Parse(reader)
	if err != nil {
		logError("Failed to parse HTML. Err: " + err.Error())
		return
	}

	var findTable func(*html.Node) bool
	findTable = func(n *html.Node) bool{
		if n.Type == html.ElementNode && n.Data == "table" && containsId(n, "proxylisttable") {
			logVerbose("Found table node with id='proxylisttable'")
			tNode = n
			return true
		}
		logVerbose("Iterate over all children of current node")
		for c := n.FirstChild; c != nil; c = c.NextSibling{
			if findTable(c) {
				logVerbose("Table already found. Stopping subsequent iterations")
				// This return prevents a unnecessary checks once the table has been found
				return true
			}
		}
		logVerbose("Table not found at this level")
		return false
	}
	logVerbose("Kickoff recursive lookup for table")
	findTable(doc) // if this is true, it found the table. TODO: Test it

	logVerbose("Table lookup finished")
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
	// TODO: Deleting all is a lazy cleanup approach. Redesign!
	logInfo("Delete all existing proxies")
	req, err := http.NewRequest(http.MethodDelete, deleteProxyBatchURL, nil)
	if err != nil {
		logError("Failed to build request. Err: " + err.Error())
		// TODO: logNotify
		return false
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logError("Failed to perform request. Err: " + err.Error())
		// TODO: logNotify
		return false
	}

	logVerbose("Response Status: " + resp.Status)
	return resp.StatusCode == http.StatusNoContent
}

func saveProxies(proxies []Proxy) bool {
	logInfo("Save batch of proxies")
	batch := Batch{ Batch: proxies }
	b, err := json.Marshal(batch)
	//b, err := json.MarshalIndent(proxies, "", " ")
	if err != nil {
		logError("Failed to Marshall list of Proxies. Err: " + err.Error())
		// TODO: logNotify
		return false
	}

	req, err := http.NewRequest(http.MethodPost, saveProxyBatchURL, bytes.NewBuffer(b))
	if err != nil {
		logError("Failed to build Post Request. Err: " + err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logError("Failed to perform Post Request. Err: " + err.Error())
		return false
	}
	defer resp.Body.Close()

	logVerbose("Response Status: " + resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logVerbose("Failed to read Response Body. Err: " + err.Error())
		return false
	}
	logVerbose("Response Body: " + string(body))
	return true
}

func main() {
	initLogger(VERBOSE)
	for {
		pageBytes, err := getPage()
		if err != nil {
			logWarning("Failed to getPage. Sleep for 2 minutes and retry.")
			time.Sleep(2 * time.Minute)
			continue
		}

		savePage(pageBytes)
		tableNode, err := findTableNode(pageBytes)
		if err != nil {
			logWarning("Failed to find table in page. Sleep for 2 minutes and retry.")

			//TODO: Implement logNotify with logs an important Error/Warning along with relevant data and sends notifications
			//logNotify("Page structure might have changed", pageBytes, err)

			time.Sleep(2 * time.Minute)
			continue
		}

		proxies := parseTable(tableNode)

		if deleteAllProxies() {
			saveProxies(proxies)
		}else {
			fmt.Printf("Could not clean up. Halting All inserts until DELETE is fixed")
		}

		logInfo("Sleeping 5 minutes...")
		ticker := time.NewTicker(10 * time.Second)
		//totalSleepSeconds := 5 * 60;
		//elapsedSeconds := 0
		go func() {
			for t := range ticker.C {
				//elapsedSeconds += 10

				


				logVerbose("Countdown: " + t.String() + " seconds...")

				//logVerbose("Countdown: " + strconv.Itoa(totalSleepSeconds - t.Second()) + " seconds...")
			}
		}()
		time.Sleep(5 * time.Minute)
		ticker.Stop()
	}
}
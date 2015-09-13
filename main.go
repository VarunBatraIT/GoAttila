// ToInsanity project main.go
package main

import (
	"github.com/mvdan/xurls"
    "github.com/PuerkitoBio/goquery"
        "io/ioutil"
		"errors"
        "net/http"
		"strings"
		"sync"
		"net/url"
		"math/rand"
		"github.com/cheggaaa/pb"
)
type hitResponse struct {  
    body string
    rawBody []byte
}
type hitRequest struct {  
    url string
    userAgent string
    params string
	method string
}
func (self *hitRequest) Initialize() {
	self.userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.130 Safari/537.36"
	self.method = "GET"
}

var targetUrl = "http://www.google.co.in/"
var numOfSoldiers = 10
var numOfBattalions = 5
var numOfTargets = 1
var bar *pb.ProgressBar
func main() {
	targets := findTarget()
	bar = pb.StartNew(numOfSoldiers * numOfBattalions  * len(targets))
	bar.Format("<.- >")
    var missionIssuedWg sync.WaitGroup
	for _, target := range targets {
		missionIssuedWg.Add(1)
		ht := hitRequest{}
		ht.Initialize()
		ht.url = target
		go deploy(ht, &missionIssuedWg)
	}
	missionIssuedWg.Wait()
	bar.FinishPrint("Victory!")
}


func deploy(ht hitRequest, missionIssuedWg *sync.WaitGroup){
    var deployWg sync.WaitGroup
	for i := 0; i < numOfBattalions; i++ {
		deployWg.Add(1)
		attack(ht,&deployWg, i)
	}
	deployWg.Wait()
	missionIssuedWg.Done()
}
func getLocalLinks(doc *goquery.Document, domain string)([]string){
	u,_ := url.Parse(domain)
	host := u.Host
	var hrefs []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link,linkExist := s.Attr("href")
    		if linkExist != false {
			if link == "#" || link == "/" || link == "" || strings.Index(link,"#") == 0 {
				return
			}
			if strings.Index(link,"/") == 0 {
				link = domain + link
			}
			u,_ = url.Parse(link)
			if host == u.Host {
				hrefs = append(hrefs,link)				
			}
		}
  })
  return hrefs
	
}
//Finds the target 
func findTarget()([]string){

	hr, err := hit(hitRequest{url: targetUrl})	
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hr.body)) 

	hrefsFromBody := xurls.Relaxed.FindAllString(doc.Text(), -1)
	if err != nil {
//		fmt.Println(doc)	
	}
	domain := targetUrl
	hrefsFromDoc := getLocalLinks(doc,domain)
	hrefs := append(hrefsFromBody,hrefsFromDoc...)
	
	//append original to the list
	hrefs = append(hrefs,domain)
	RemoveDuplicatesStringSlice(&hrefs)
	ShuffleStringSlice(hrefs)
	var targets []string
	for i := 0; i < numOfTargets; i++ {
		url := hrefs[len(hrefs) - 1]
		if strings.Index(url,"@") > 0 {
			i--
			continue
		}
		targets = append(targets,url)
		hrefs = hrefs[:len(hrefs)-1]
	}
	RemoveDuplicatesStringSlice(&targets)
	return targets 
}

func hit(ht hitRequest)(hitResponse, error){
		
		client := &http.Client{}
        req, err := http.NewRequest(ht.method, targetUrl, nil)    
	    if err != nil {
				return hitResponse{},errors.New("Can't hit it")
        }
		if ht.userAgent != "" {
			req.Header.Set("User-Agent", ht.userAgent)
		}
        resp, err := client.Do(req)
        if err != nil {
				return hitResponse{},errors.New("Can't hit it")
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)

        if err != nil {
				return hitResponse{}, errors.New("Can't hit it")
        }
		return hitResponse{string(body),body}, nil
}


func attack(ht hitRequest,deployWg *sync.WaitGroup, numBattalion int){

    var wg sync.WaitGroup
	messages := make(chan int)
	for i := 0; i < numOfSoldiers; i++ {
		wg.Add(1)
		go kill(ht, messages, &wg, i, numBattalion)
	}
	wg.Wait()
	deployWg.Done()

}
func kill(ht hitRequest,messages chan int, wg *sync.WaitGroup, numSoldier int, numBattalion int){
	    hit(ht)
		bar.Increment()
		wg.Done()
		messages <- numSoldier
}

func ShuffleStringSlice(a []string) {
    for i := range a {
        j := rand.Intn(i + 1)
        a[i], a[j] = a[j], a[i]
    }
}
func ShuffleIntegerSlice(a []int) {
    for i := range a {
        j := rand.Intn(i + 1)
        a[i], a[j] = a[j], a[i]
    }
}
func RemoveDuplicatesStringSlice(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

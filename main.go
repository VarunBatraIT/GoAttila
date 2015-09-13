// ToInsanity project main.go
package main

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"github.com/cheggaaa/pb"
	"golang.org/x/net/html"
)

type hitResponse struct {
	body    string
	rawBody []byte
}
type hitRequest struct {
	url       string
	userAgent string
	params    string
	method    string
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
	bar = pb.StartNew(numOfSoldiers * numOfBattalions * len(targets))
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
// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
    // Iterate over all of the Token's attributes until we find an "href"
    for _, a := range t.Attr {
        if a.Key == "href" {
            href = a.Val
            ok = true
        }
    }

    // "bare" return will return the variables (ok, href) as defined in
    // the function definition
    return
}

func deploy(ht hitRequest, missionIssuedWg *sync.WaitGroup) {
	var deployWg sync.WaitGroup
	for i := 0; i < numOfBattalions; i++ {
		deployWg.Add(1)
		attack(ht, &deployWg, i)
	}
	deployWg.Wait()
	missionIssuedWg.Done()
}
func getLocalLinks(doc string, domain string) []string {
	u,_:= url.Parse(domain)
	host := u.Host
	var hrefs []string
	z := html.NewTokenizer(strings.NewReader(doc))

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return hrefs
		case tt == html.StartTagToken:
			t := z.Token()
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			// Extract the href value, if there is one
            ok, href := getHref(t)
            if !ok {
                continue
            }
			if strings.Index(href,"/") == 0 {
				href = domain + href
			}
			urlParsed,_ := url.Parse(href)
			if  urlParsed.Host == host{
				hrefs = append(hrefs,href)				
			}
		}
	}
	return hrefs
}

//Finds the target
func findTarget() []string {
	hr, _ := hit(hitRequest{url: targetUrl})
	domain := targetUrl
	hrefs := getLocalLinks(hr.body, domain)

	//append original to the list
	hrefs = append(hrefs, domain)
	RemoveDuplicatesStringSlice(&hrefs)
	ShuffleStringSlice(hrefs)
	var targets []string
	for i := 0; i < numOfTargets; i++ {
		url := hrefs[len(hrefs)-1]
		if strings.Index(url, "@") > 0 {
			i--
			continue
		}
		targets = append(targets, url)
		hrefs = hrefs[:len(hrefs)-1]
	}
	RemoveDuplicatesStringSlice(&targets)
	return targets
}

func hit(ht hitRequest) (hitResponse, error) {

	client := &http.Client{}
	req, err := http.NewRequest(ht.method, targetUrl, nil)
	if err != nil {
		return hitResponse{}, errors.New("Can't hit it")
	}
	if ht.userAgent != "" {
		req.Header.Set("User-Agent", ht.userAgent)
	}
	resp, err := client.Do(req)
	if err != nil {
		return hitResponse{}, errors.New("Can't hit it")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return hitResponse{}, errors.New("Can't hit it")
	}
	return hitResponse{string(body), body}, nil
}

func attack(ht hitRequest, deployWg *sync.WaitGroup, numBattalion int) {

	var wg sync.WaitGroup
	messages := make(chan int)
	for i := 0; i < numOfSoldiers; i++ {
		wg.Add(1)
		go kill(ht, messages, &wg, i, numBattalion)
	}
	wg.Wait()
	deployWg.Done()

}
func kill(ht hitRequest, messages chan int, wg *sync.WaitGroup, numSoldier int, numBattalion int) {
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

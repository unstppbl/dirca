package dirca

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
)

// Dirber struct
type Dirber struct {
	transport *http.Transport
	opts      *Options
	mu        *sync.RWMutex
}

// NewDirber initializes dirber and returns it
func NewDirber(options *Options) (dirber *Dirber, err error) {
	dirber = &Dirber{}
	exts, err := getStringSetFromFile(options.ExtensionsPath)
	if err != nil {
		return nil, err
	}
	wordlist, err := getStringSetFromFile(options.WordlistPath)
	if err != nil {
		return nil, err
	}
	codes, err := getIntSetFromFile(options.StatusCodesPath)
	if err != nil {
		return nil, err
	}
	dirber.opts = options
	dirber.opts.Extensions = exts
	dirber.opts.Wordlist = wordlist
	dirber.opts.StatusCodes = codes
	dirber.opts.wordlistLength = len(wordlist)
	dirber.mu = new(sync.RWMutex)
	httpTransport, err := NewHTTPTransport(options)
	if err != nil {
		return nil, err
	}
	dirber.transport = httpTransport
	return dirber, nil
}

// Setup is the setup implementation of gobusterdir
func (d *Dirber) Setup(URL string) error {
	sc, l, err := d.makeRequest(URL)
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %v", URL, err)
	}
	log.Infof("[*] URL is online: status code - %d, page length - %d", *sc, *l)
	guid := uuid.New()
	url := fmt.Sprintf("%s%s", URL, guid)
	wildcardResp, len, err := d.makeRequest(url)
	if err != nil {
		return err
	}

	if d.opts.StatusCodes[*wildcardResp] {
		log.Printf("[-] Wildcard response found: %s => %d", url, *wildcardResp)
		d.opts.WildcardFound = true
		if d.opts.IgnoreWildcard {
			log.Print("[-] Ignoring wilcard responses")
			d.opts.WildcardRespLen = len
			return nil
		}
		return fmt.Errorf("[-] Wildcard response found: %s => %d", url, *wildcardResp)
	}

	return nil
}

// Process is the process implementation of gobusterdir
func (d *Dirber) process(word string, URL string) ([]Result, error) {
	// Try the DIR first
	url := fmt.Sprintf("%s%s", URL, word)
	dirResp, dirSize, err := d.makeRequest(url)
	if err != nil {
		return nil, err
	}
	var results []Result
	if dirResp != nil {
		results = append(results, Result{
			Entity: word,
			Status: *dirResp,
			Size:   *dirSize,
		})
	}

	// Follow up with files using each ext.
	if d.opts.SearchExtensions {
		for ext := range d.opts.Extensions {
			file := strings.Replace(ext, "%", word, 1)
			url = fmt.Sprintf("%s%s", URL, file)
			fileResp, fileSize, err := d.makeRequest(url)
			if err != nil {
				return nil, err
			}

			if fileResp != nil {
				results = append(results, Result{
					Entity: file,
					Status: *fileResp,
					Size:   *fileSize,
				})
			}
		}
	}
	return results, nil
}

func (d *Dirber) worker(requestsIssued *int, url string, words chan string, results chan Result, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	for word := range words {
		d.incrementRequests(requestsIssued)
		res, err := d.process(word, url)
		if err != nil {
			errors <- err
			continue
		} else {
			for _, r := range res {
				results <- r
			}
		}
	}
}

// Scan scans the URL
func (d *Dirber) Scan(URL string) (results []Result, err error) {
	// Check if host is available and does not have wildcard responses
	if err := d.Setup(URL); err != nil {
		return nil, err
	}
	// Workers wait group and channel
	workersWG := &sync.WaitGroup{}
	wordsChan := make(chan string, len(d.opts.Wordlist))
	// Processors wait group and channels
	processorsWG := &sync.WaitGroup{}
	resultsChan := make(chan Result)
	errorsChan := make(chan error)

	// Run workers
	requestsIssued := new(int)
	for i := 1; i <= d.opts.Threads; i++ {
		workersWG.Add(1)
		go d.worker(requestsIssued, URL, wordsChan, resultsChan, errorsChan, workersWG)
	}
	// Print progress
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go d.progressWorker(ctx, requestsIssued)
	// Run results and errors workers
	processorsWG.Add(2)
	go d.processResults(resultsChan, &results, processorsWG)
	go d.processErrors(errorsChan, processorsWG)

	// Send words to workers
	for w := range d.opts.Wordlist {
		wordsChan <- w
	}
	close(wordsChan)
	workersWG.Wait()
	close(resultsChan)
	close(errorsChan)
	processorsWG.Wait()
	return results, nil
}

// Results worker
func (d *Dirber) processResults(resultsChan chan Result, results *[]Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range resultsChan {
		if res.Status == http.StatusOK {
			if d.opts.WildcardFound && d.opts.IgnoreWildcard {
				if res.Size == *d.opts.WildcardRespLen {
					continue
				}
			}
			d.ClearProgress()
			log.Infof("[*] %d --> /%s", res.Status, res.Entity)
			*results = append(*results, res)
		}
	}
}

// Errors worker
func (d *Dirber) processErrors(errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	for err := range errors {
		d.ClearProgress()
		log.Errorf("[!] %s", err)
	}
}

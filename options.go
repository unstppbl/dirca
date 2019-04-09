package dirca

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Options helds all options that can be passed to libgobuster
type Options struct {
	Extensions       map[string]bool
	StatusCodes      map[int]bool
	Wordlist         map[string]bool
	Threads          int
	Proxy            string
	Cookies          string
	Timeout          time.Duration
	FollowRedirect   bool
	IncludeLength    bool
	InsecureSSL      bool
	SearchExtensions bool
	wordlistLength   int
	ExtensionsPath   string
	WordlistPath     string
	StatusCodesPath  string
	IgnoreWildcard   bool
	WildcardFound    bool
	WildcardRespLen  *int64
}

func getStringSetFromFile(fileName string) (set map[string]bool, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	set = map[string]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore commented lines
		if strings.Contains(line, "#") {
			continue
		}
		set[line] = true
	}
	if scanner.Err() != nil {
		return nil, err
	}
	return set, nil
}

func getIntSetFromFile(fileName string) (set map[int]bool, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	set = map[int]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		statusCode, err := strconv.Atoi(line)
		if err != nil {
			log.Errorln(err)
			continue
		}
		set[statusCode] = true
	}
	if scanner.Err() != nil {
		return nil, err
	}
	return set, nil
}

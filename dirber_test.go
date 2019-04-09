package dirber

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDirber(t *testing.T) {
	url := "http://cybersec.kz/"
	options := &Options{
		Threads:          30,
		Timeout:          time.Second * 5,
		FollowRedirect:   true,
		IncludeLength:    true,
		InsecureSSL:      true,
		IgnoreWildcard:   true,
		SearchExtensions: false,
		//Proxy:            "socks5://13.69.58.220:1080",
		ExtensionsPath:  "../exts.txt",
		WordlistPath:    "../fuzz.txt",
		StatusCodesPath: "../codes.txt",
	}
	dirber, err := NewDirber(options)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !strings.HasSuffix(url, "/") {
		url = fmt.Sprintf("%s/", url)
	}
	res, err := dirber.Scan(url)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	dirber.ClearProgress()
	for _, r := range res {
		fmt.Printf("[!!!] %s\n", r.Entity)
	}
}

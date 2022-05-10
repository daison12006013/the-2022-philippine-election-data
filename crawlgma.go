// Author: Daison Cariño
// You may copy this file and place it inside
// your forked github.com/lucidfy/lucid,
// under this path "app/commands/"
//
// Go  to registrar/commands.go and inject this
// commands.CrawlGma().Command,
//
// Then execute the bash command `run`  under your lucid folder
//     ./run crawlgma --procedure batch --counter 100
//     ./run crawlgma --procedure hourly --counter 50

package commands

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/lucidfy/lucid/pkg/facade/logger"
	"github.com/lucidfy/lucid/pkg/facade/path"
	"github.com/lucidfy/lucid/pkg/functions/php"
	cli "github.com/urfave/cli/v2"
)

const FOLDER = "./../election-2022-data-transparency/"

type CrawlGmaCommand struct {
	Command *cli.Command
}

func CrawlGma() *CrawlGmaCommand {
	var cc CrawlGmaCommand
	cc.Command = &cli.Command{
		Name:    "crawlgma",
		Aliases: []string{},
		Usage:   "",
		Action:  cc.Handle,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "procedure",
				Value: "batch",
				Usage: `Provider either "batch" or "hourly"`,
			},
			&cli.IntFlag{
				Name:  "counter",
				Value: 100,
				Usage: `Provide the loop counter, ie "100", better to lessen it to avoid request leaks!`,
			},
		},
	}
	return &cc
}

func (cc *CrawlGmaCommand) Handle(c *cli.Context) error {
	procedure := c.String("procedure")
	maxCounter := c.Int("counter")

	urlFormat := "{baseUrl}/{counter}/{region}.json"
	baseUrl := "https://e22vh.gmanetwork.com/" + procedure
	regions := []string{
		"PRESIDENT_PHILIPPINES",
		"VICE_PRESIDENT_PHILIPPINES",
		"SENATOR_PHILIPPINES",
		"NATIONAL_CAPITAL_REGION",
		"CORDILLERA_ADMINISTRATIVE_REGION",
		"REGION_I",
		"REGION_II",
		"REGION_III",
		"REGION_IV-A",
		"REGION_IV-B",
		"REGION_V",
		"REGION_VI",
		"REGION_VII",
		"REGION_VIII",
		"REGION_IX",
		"REGION_X",
		"REGION_XI",
		"REGION_XII",
		"REGION_XIII",
		"BARMM",
		"OAV",
	}

	var contentsChan = make(chan map[string]interface{})
	var wg sync.WaitGroup

	for _, r := range regions {
		for i := 1; i <= maxCounter; i++ {
			u := urlFormat
			u = strings.ReplaceAll(u, "{baseUrl}", baseUrl)
			u = strings.ReplaceAll(u, "{counter}", strconv.Itoa(i))
			u = strings.ReplaceAll(u, "{region}", r)

			if php.FileExists(path.Load().BasePath(FOLDER + "/" + procedure + "/" + strconv.Itoa(i) + "/" + filepath.Base(u))) {
				continue
			}

			wg.Add(1)
			go cUrlWebsite(u, i, contentsChan, &wg)
		}
	}

	go func() {
		wg.Wait()
		close(contentsChan)
	}()

	for resp := range contentsChan {
		pathToRegionFolder := FOLDER + "/" + procedure + "/" + strconv.Itoa(resp["batch_id"].(int))

		if strings.Contains(resp["content"].(string), "File Not Found!") {
			continue
		}

		php.Mkdir(pathToRegionFolder, 0755, true)

		php.FilePutContents(
			path.Load().BasePath(pathToRegionFolder+"/"+filepath.Base(resp["url"].(string))),
			resp["content"].(string),
			0755,
		)
	}

	return nil
}

func cUrlWebsite(link string, counter int, contentsChan chan map[string]interface{}, wg *sync.WaitGroup) {
	logger.Info("Fetching : "+strconv.Itoa(counter), filepath.Base(link))

	defer wg.Done()

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authority", "e22vh.gmanetwork.com")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Origin", "https://www.gmanetwork.com")
	req.Header.Set("Referer", "https://www.gmanetwork.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Gpc", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("err", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Err readAll", err)
	}

	contentsChan <- map[string]interface{}{
		"batch_id": counter,
		"url":      link,
		"content":  string(body),
	}
}

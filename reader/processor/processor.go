// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"miniflux.app/http/client"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/rewrite"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/reader/scraper"
	"miniflux.app/storage"
)

// ProcessFeedEntries downloads original web page for entries and apply filters.
func ProcessFeedEntries(store *storage.Storage, feed *model.Feed) {
	// TODO: maybe the score extractor should be called by ProcessFeedEntries's caller
	err := ProcessScoreExtractor(feed)
	if err != nil {
		logger.Debug("[Feed #%d] Error extracting score: %v", feed.ID, err)
	}

	var wg sync.WaitGroup
	wg.Add(len(feed.Entries))

	for _, entry := range feed.Entries {
		go func(entry *model.Entry) {
			defer wg.Done()

			logger.Debug("[Feed #%d] Processing entry %s", feed.ID, entry.URL)

			if feed.Crawler {
				if !store.EntryURLExists(feed.ID, entry.URL) {
					content, err := scraper.Fetch(entry.URL, feed.ScraperRules, feed.UserAgent)
					if err != nil {
						logger.Error(`[Filter] Unable to crawl this entry: %q => %v`, entry.URL, err)
					} else if content != "" {
						// We replace the entry content only if the scraper doesn't return any error.
						entry.Content = content
					}
				}
			}

			entry.Content = rewrite.Rewriter(entry.URL, entry.Content, feed.RewriteRules)

			// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered.
			entry.Content = sanitizer.Sanitize(entry.URL, entry.Content)
		}(entry)
	}
	wg.Wait()
}

// TODO: Move this to a different file/package
// ProcessScoreExtractor tries to obtain the score for the feed entries
func ProcessScoreExtractor(feed *model.Feed) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("Panic while processing score extractor: %v", r))
		}
	}()

	if feed.ScoreExtractor == "reddit" {
		// Build id string
		ids := make([]string, 0, len(feed.Entries))
		idEntryMap := make(map[string]int)
		for i, entry := range feed.Entries {
			if entry.OriginalID != "" {
				ids = append(ids, entry.OriginalID)
				idEntryMap[entry.OriginalID] = i
			}
		}
		idParam := strings.Join(ids, ",")

		url := fmt.Sprintf("https://reddit.com/api/info.json?id=%s", idParam)
		clt := client.New(url)
		response, err := clt.Get()

		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return err
		}

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)

		if err != nil {
			return err
		}

		// The values I want are in data.children.[0..n].data.[ups, name]
		posts := result["data"].(map[string]interface{})["children"].([]interface{})

		for _, post := range posts {
			data := post.(map[string]interface{})["data"].(map[string]interface{})
			id := data["name"].(string)
			score := data["ups"].(float64)

			index, prs := idEntryMap[id]

			if prs {
				feed.Entries[index].Score = int64(score)
			}
		}
	}

	return nil
}

// ProcessEntryWebPage downloads the entry web page and apply rewrite rules.
func ProcessEntryWebPage(entry *model.Entry) error {
	content, err := scraper.Fetch(entry.URL, entry.Feed.ScraperRules, entry.Feed.UserAgent)
	if err != nil {
		return err
	}

	content = rewrite.Rewriter(entry.URL, content, entry.Feed.RewriteRules)
	content = sanitizer.Sanitize(entry.URL, content)

	if content != "" {
		entry.Content = content
	}

	return nil
}

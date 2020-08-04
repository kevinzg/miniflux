package ui // import "miniflux.app/ui"

import (
	"net/http"
	"strconv"

	"github.com/gorilla/feeds"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/model"
	"miniflux.app/storage"
)

// TODO: use ui/pagination?
const nbItemsPerPage = 100

func (h *handler) showRSSEntriesPage(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := h.store.AnonFeedByID(feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if feed == nil {
		html.NotFound(w, r)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := storage.NewAnonymousQueryBuilder(h.store)
	builder.WithFeedID(feed.ID)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection("desc")
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	gorillaFeed := &feeds.Feed{
		Title:       feed.Title,
		Link:        &feeds.Link{Href: feed.SiteURL},
		Description: "RSS from Miniflux reader",
		Updated:     feed.CheckedAt,
	}

	for _, item := range entries {
		entry := &feeds.Item{
			Id:      strconv.FormatInt(item.ID, 10),
			Title:   item.Title,
			Link:    &feeds.Link{Href: item.URL},
			Created: item.Date,
			Updated: item.Date,
			Content: item.Content,
		}

		// TODO: `item.Author` could be an email
		if item.Author != "" {
			entry.Author = &feeds.Author{
				Name: item.Author,
			}
		}

		if len(item.Enclosures) > 1 {
			entry.Enclosure = &feeds.Enclosure{
				Url:    item.Enclosures[0].URL,
				Length: strconv.FormatInt(item.Enclosures[0].Size, 10),
				Type:   item.Enclosures[0].MimeType,
			}
		}

		gorillaFeed.Add(entry)
	}

	// TODO: Add cache headers?
	w.Header().Set("Content-Type", "application/atom+xml")
	gorillaFeed.WriteAtom(w)
}

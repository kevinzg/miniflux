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
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// TODO: Complete fields?
	gorillaFeed := &feeds.Feed{
		Title: feed.Title,
		Link:  &feeds.Link{Href: feed.SiteURL},
	}

	// TODO: Add author, enclosures, etc
	for _, item := range entries {
		gorillaFeed.Add(&feeds.Item{
			Id:      strconv.FormatInt(item.ID, 10),
			Title:   item.Title,
			Link:    &feeds.Link{Href: item.URL},
			Created: item.Date,
			Updated: item.Date,
			Content: item.Content,
		})
	}

	// TODO: Missing headers?
	gorillaFeed.WriteAtom(w)
}

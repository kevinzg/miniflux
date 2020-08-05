Miniflux 2
==========
[![GoDoc](https://godoc.org/miniflux.app?status.svg)](https://godoc.org/miniflux.app)

Miniflux is a minimalist and opinionated feed reader.
Official website: <https://miniflux.app>

My modifications
----------------
- Add a public RSS endpoint: `https://miniflux/rss/<feed-id>`
- Add `resolve_redirects` rewriter.
    - PR: <https://github.com/miniflux/miniflux/pull/676>
- Add score to feed entries and a score extractor for Reddit.
    - Run `miniflux -fx-migrate` to add the columns.
    - Manually update the `score_extractor` column to `reddit` for the Reddit feeds.
    - Reddit entries will show the upvotes and be sorted by it.

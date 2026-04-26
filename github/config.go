package github

import "time"

// CacheTTL は starred repos キャッシュの有効期間。
const CacheTTL = time.Hour

// SearchLimit はグローバル検索の最大件数。
const SearchLimit = 30

package structures

type DocumentIndex struct {
	NodeMap   map[int64]int64
}

type CacheEntry struct {
	Results []int64
	Found   int
}

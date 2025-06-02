package structures

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	ngramSize      = 3
	combinedRegexp = regexp.MustCompile("[^a-z0-9]+")
)

type ngramRecord struct {
	ngram  string
	docIDs []int64
}

type docRecord struct {
	docID int64
	text  string
}

type Index struct {
	ngrams []ngramRecord
	docs   []docRecord
	mutex  sync.RWMutex
}

func NewIndex() *Index {
	return &Index{
		ngrams: make([]ngramRecord, 0),
		docs:   make([]docRecord, 0),
	}
}

// -------------------------------------------------------------------
// Adding Documents
// -------------------------------------------------------------------

// AddDocument appends an unsorted entry for the doc in memory.
// We'll rely on Optimize() to sort/deduplicate everything later.
func (idx *Index) AddDocument(id int64, text string) {
	text = normalizeString(text)
	ngrams := generateNGrams(text, ngramSize)

	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Append docRecord, not inserted in sorted order yet.
	idx.docs = append(idx.docs, docRecord{
		docID: id,
		text:  text,
	})

	// For each n-gram, we do a naive append (we do NOT yet keep them in sorted order).
	for _, ng := range ngrams {
		idx.ngrams = append(idx.ngrams, ngramRecord{
			ngram:  ng,
			docIDs: []int64{id},
		})
	}
}

// Optimize sorts and deduplicates the internal slices.
//
// 1) Sort idx.docs by docID and remove duplicates (if any).
// 2) Group idx.ngrams by ngram and merge docIDs together, then sort/deduplicate docIDs.
func (idx *Index) Optimize() {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Step 1: sort docs by docID
	sort.Slice(idx.docs, func(i, j int) bool {
		return idx.docs[i].docID < idx.docs[j].docID
	})
	// optional deduplicate docIDs if your data might have duplicates
	idx.docs = deduplicateDocs(idx.docs)

	// Step 2: sort the ngrams by ngram string
	sort.Slice(idx.ngrams, func(i, j int) bool {
		return idx.ngrams[i].ngram < idx.ngrams[j].ngram
	})

	// We'll merge consecutive records with the same ngram.
	merged := make([]ngramRecord, 0, len(idx.ngrams))
	for i := 0; i < len(idx.ngrams); {
		cur := idx.ngrams[i]
		j := i + 1
		for j < len(idx.ngrams) && idx.ngrams[j].ngram == cur.ngram {
			// merge docIDs
			cur.docIDs = append(cur.docIDs, idx.ngrams[j].docIDs...)
			j++
		}
		// sort & deduplicate docIDs
		sort.Slice(cur.docIDs, func(a, b int) bool {
			return cur.docIDs[a] < cur.docIDs[b]
		})
		cur.docIDs = deduplicateInt64(cur.docIDs)

		merged = append(merged, cur)
		i = j
	}
	idx.ngrams = merged
}

// deduplicateDocs merges duplicates if docID is repeated
func deduplicateDocs(docs []docRecord) []docRecord {
	if len(docs) <= 1 {
		return docs
	}
	out := docs[:1]
	for i := 1; i < len(docs); i++ {
		if docs[i].docID != out[len(out)-1].docID {
			out = append(out, docs[i])
		}
	}
	return out
}

// deduplicateInt64 removes duplicates in a sorted []int64.
func deduplicateInt64(arr []int64) []int64 {
	if len(arr) <= 1 {
		return arr
	}
	out := arr[:1]
	for i := 1; i < len(arr); i++ {
		if arr[i] != out[len(out)-1] {
			out = append(out, arr[i])
		}
	}
	return out
}

// -------------------------------------------------------------------
// Searching
// -------------------------------------------------------------------

func (idx *Index) Search(query string, maxDistance int) []int64 {
	normalizedQuery := normalizeString(query)
	qNgrams := generateNGrams(normalizedQuery, ngramSize)
	if len(qNgrams) == 0 {
		return nil
	}
	threshold := len(qNgrams) / 2

	candidateCounts := make(map[int64]int)

	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	for _, qng := range qNgrams {
		if i := idx.findNgram(qng); i >= 0 {
			for _, docID := range idx.ngrams[i].docIDs {
				candidateCounts[docID]++
			}
		}
	}

	results := make([]int64, 0, len(candidateCounts))
	for docID, count := range candidateCounts {
		if count >= threshold {
			text := idx.findDocText(docID)
			if text != "" && levenshteinDistanceWithinMax(normalizedQuery, text, maxDistance) {
				results = append(results, docID)
			}
		}
	}
	return results
}

// findNgram does a binary search in idx.ngrams for ngram=key.
// Returns the index or -1 if not found.
func (idx *Index) findNgram(key string) int {
	lo, hi := 0, len(idx.ngrams)-1
	for lo <= hi {
		mid := (lo + hi) >> 1
		if idx.ngrams[mid].ngram < key {
			lo = mid + 1
		} else if idx.ngrams[mid].ngram > key {
			hi = mid - 1
		} else {
			return mid
		}
	}
	return -1
}

// findDocText does a binary search in idx.docs for docID=id.
// Returns the text or "" if not found.
func (idx *Index) findDocText(id int64) string {
	lo, hi := 0, len(idx.docs)-1
	for lo <= hi {
		mid := (lo + hi) >> 1
		if idx.docs[mid].docID < id {
			lo = mid + 1
		} else if idx.docs[mid].docID > id {
			hi = mid - 1
		} else {
			return idx.docs[mid].text
		}
	}
	return ""
}

// -------------------------------------------------------------------
// Custom Binary Serialization
// -------------------------------------------------------------------
/*
Format (high-level):

1) uint64 = number of n-gram records
   then for each:
    1.1) uint64 = length of the n-gram (L)
    1.2) [L] bytes = n-gram string
    1.3) uint64 = number of docIDs (D)
    1.4) D x int64 = docIDs

2) uint64 = number of docRecords
   then for each:
    2.1) int64 = docID
    2.2) uint64 = length of text (T)
    2.3) [T] bytes = text
*/

func (idx *Index) Save(w io.Writer) error {
	idx.Optimize()
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	writer := bufio.NewWriter(w)

	// 1) Write number of ngramRecords
	nCount := uint64(len(idx.ngrams))
	if err := binary.Write(writer, binary.LittleEndian, nCount); err != nil {
		return err
	}
	// For each ngramRecord
	for _, rec := range idx.ngrams {
		// 1.1) length of the n-gram
		ngBytes := []byte(rec.ngram)
		if err := binary.Write(writer, binary.LittleEndian, uint64(len(ngBytes))); err != nil {
			return err
		}
		// 1.2) n-gram bytes
		if _, err := writer.Write(ngBytes); err != nil {
			return err
		}
		// 1.3) number of docIDs
		dCount := uint64(len(rec.docIDs))
		if err := binary.Write(writer, binary.LittleEndian, dCount); err != nil {
			return err
		}
		// 1.4) docIDs
		for _, d := range rec.docIDs {
			if err := binary.Write(writer, binary.LittleEndian, d); err != nil {
				return err
			}
		}
	}
	// 2) Write number of docRecords
	docsCount := uint64(len(idx.docs))
	if err := binary.Write(writer, binary.LittleEndian, docsCount); err != nil {
		return err
	}
	// For each docRecord
	for _, dr := range idx.docs {
		// 2.1) docID
		if err := binary.Write(writer, binary.LittleEndian, dr.docID); err != nil {
			return err
		}
		// 2.2) length of text
		txtBytes := []byte(dr.text)
		if err := binary.Write(writer, binary.LittleEndian, uint64(len(txtBytes))); err != nil {
			return err
		}
		// 2.3) text bytes
		if _, err := writer.Write(txtBytes); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (idx *Index) Load(data []byte) error {
	bytesReader := bytes.NewReader(data)
	return idx.LoadFromReader(bytesReader)
}

func (idx *Index) LoadFromReader(r io.Reader) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	reader := bufio.NewReaderSize(r, 256*1024) // 256KB buffer

	var nCount uint64
	if err := binary.Read(reader, binary.LittleEndian, &nCount); err != nil {
		return err
	}

	newNgrams := make([]ngramRecord, 0, nCount)
	for i := uint64(0); i < nCount; i++ {
		// 1.1) ngram length
		var l uint64
		if err := binary.Read(reader, binary.LittleEndian, &l); err != nil {
			return err
		}
		// 1.2) read ngram
		ngBytes := make([]byte, l)
		if _, err := io.ReadFull(reader, ngBytes); err != nil {
			return err
		}
		// 1.3) number of docIDs
		var dCount uint64
		if err := binary.Read(reader, binary.LittleEndian, &dCount); err != nil {
			return err
		}
		docIDs := make([]int64, dCount)
		for j := uint64(0); j < dCount; j++ {
			if err := binary.Read(reader, binary.LittleEndian, &docIDs[j]); err != nil {
				return err
			}
		}
		newNgrams = append(newNgrams, ngramRecord{
			ngram:  string(ngBytes),
			docIDs: docIDs,
		})
	}

	// 2) read number of docRecords
	var docsCount uint64
	if err := binary.Read(reader, binary.LittleEndian, &docsCount); err != nil {
		return err
	}

	newDocs := make([]docRecord, 0, docsCount)
	for i := uint64(0); i < docsCount; i++ {
		// 2.1) docID
		var dID int64
		if err := binary.Read(reader, binary.LittleEndian, &dID); err != nil {
			return err
		}
		// 2.2) text length
		var tLen uint64
		if err := binary.Read(reader, binary.LittleEndian, &tLen); err != nil {
			return err
		}
		// 2.3) text bytes
		txtBytes := make([]byte, tLen)
		if _, err := io.ReadFull(reader, txtBytes); err != nil {
			return err
		}
		newDocs = append(newDocs, docRecord{
			docID: dID,
			text:  string(txtBytes),
		})
	}

	// Reassign
	idx.ngrams = newNgrams
	idx.docs = newDocs
	return nil
}

func (idx *Index) LoadFromFile(file io.ReaderAt, offset int64, size uint64) error {
	reader := io.NewSectionReader(file, offset, int64(size))
	return idx.LoadFromReader(reader)
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func normalizeString(s string) string {
	s = strings.ToLower(s)
	s = combinedRegexp.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func generateNGrams(s string, n int) []string {
	if len(s) < n {
		return nil
	}
	out := make([]string, 0, len(s)-n+1)
	for i := 0; i <= len(s)-n; i++ {
		out = append(out, s[i:i+n])
	}
	return out
}

// Levenshtein distance with early exit
func levenshteinDistanceWithinMax(a, b string, max int) bool {
	al := len(a)
	bl := len(b)
	if al == 0 {
		return bl <= max
	}
	if bl == 0 {
		return al <= max
	}
	if al < bl {
		a, b = b, a
		al, bl = bl, al
	}
	prevRow := make([]int, bl+1)
	for j := 0; j <= bl; j++ {
		prevRow[j] = j
	}
	for i := 1; i <= al; i++ {
		curRow := make([]int, bl+1)
		curRow[0] = i
		minVal := i
		for j := 1; j <= bl; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curRow[j] = min3(
				prevRow[j]+1,
				curRow[j-1]+1,
				prevRow[j-1]+cost,
			)
			if curRow[j] < minVal {
				minVal = curRow[j]
			}
		}
		if minVal > max {
			return false
		}
		prevRow = curRow
	}
	return prevRow[bl] <= max
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

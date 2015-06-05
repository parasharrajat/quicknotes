package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/dustin/go-humanize"
	"github.com/kjk/lzmadec"
	"github.com/kjk/stackoverflow"
	"github.com/kjk/u"
)

var (
	dataDir           string
	posts             map[int]*PostChange
	historyTypeCounts map[int]int
)

type PostChange struct {
	postID int
	userID int
	typ    int
	val    string
	next   *PostChange
}

func init() {
	dataDir = u.ExpandTildeInPath("~/data/import_stack_overflow")
	posts = make(map[int]*PostChange)
	historyTypeCounts = make(map[int]int)
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatalf("err: %s\n", err)
	}
}

func toBytes(n uint64) string {
	return humanize.Bytes(n)
}

func dumpMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("Alloc      : %s\n", toBytes(ms.Alloc))
	fmt.Printf("HeapAlloc  : %s\n", toBytes(ms.HeapAlloc))
	fmt.Printf("HeapSys    : %s\n", toBytes(ms.HeapSys))
	fmt.Printf("HeapInuse  : %s\n", toBytes(ms.HeapInuse))
	fmt.Printf("HeapObjects: %d\n", ms.HeapObjects)
}

func isValidType(typ int) bool {
	switch typ {
	case stackoverflow.HistoryInitialTitle:
	case stackoverflow.HistoryInitialBody:
	case stackoverflow.HistoryInitialTags:
	case stackoverflow.HistoryEditTitle:
	case stackoverflow.HistoryEditBody:
	case stackoverflow.HistoyrEditTags:
	case stackoverflow.HistoryRollbackTitle:
	case stackoverflow.HistoryRollbackBody:
	case stackoverflow.HistoryRollbackTags:
		return true
	}
	return false
}

func postHistoryToPostChange(ph *stackoverflow.PostHistory) *PostChange {
	var pc PostChange
	if !isValidType(ph.PostHistoryTypeID) {
		return nil
	}
	pc.postID = ph.PostID
	pc.userID = ph.UserID
	pc.typ = ph.PostHistoryTypeID
	return &pc
}

func getHistoryReader(site string) *stackoverflow.Reader {
	archiveFileName := site + ".stackexchange.com.7z"
	archiveFilePath := filepath.Join(dataDir, archiveFileName)
	archive, err := lzmadec.NewArchive(archiveFilePath)
	fatalIfErr(err)
	r, err := archive.GetFileReader("PostHistory.xml")
	fatalIfErr(err)
	hr, err := stackoverflow.NewPostHistoryReader(r)
	fatalIfErr(err)
	return hr
}

func dumpCounts(m map[int]int) {
	max := 0
	for k := range m {
		if k > max {
			max = k
		}
	}
	fmt.Print("History type counts:\n")
	for i := 0; i <= max; i++ {
		if count, ok := m[i]; ok {
			fmt.Printf("type: %d, count: %d\n", i, count)
		}
	}
}

func main() {
	hr := getHistoryReader("academia")
	n := 0
	for hr.Next() {
		n++
		historyTypeCounts[hr.PostHistory.PostHistoryTypeID]++
		pc := postHistoryToPostChange(&hr.PostHistory)
		if pc == nil {
			continue
		}
		pc.next = posts[pc.postID]
		posts[pc.postID] = pc
	}
	err := hr.Err()
	fatalIfErr(err)
	fmt.Printf("%d history entries, %d posts\n", n, len(posts))
	dumpCounts(historyTypeCounts)
	dumpMemStats()
}
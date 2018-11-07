package gocore

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Stat comment
type Stat struct {
	mu         sync.RWMutex
	key        string
	parent     *Stat
	children   map[string]*Stat
	firstNanos int64
	lastNanos  int64
	minNanos   int64
	maxNanos   int64
	total      int64
	count      int64
	firstTime  time.Time
	lastTime   time.Time
}

var (
	initTime = time.Now().UTC()
	rootItem = &Stat{
		key:      "root",
		children: make(map[string]*Stat),
	}
)

func handleStats(w http.ResponseWriter, r *http.Request) {
	rootItem.printStatisticsHTML(w, "")
}

// StartServer comment
func StartServer(addr string) {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/js/", fs)
	http.Handle("/css/", fs)
	http.HandleFunc("/stats", handleStats)

	var err = http.ListenAndServe(addr, nil)

	if err != nil {
		log.Panicln("Server failed starting. Error: %s", err)
	}
}

// NewStat comment
func NewStat(key string) *Stat {

	parent := rootItem

	parent.mu.Lock()
	defer parent.mu.Unlock()

	s, ok := parent.children[key]
	if !ok {
		s = &Stat{
			key:      key,
			parent:   parent,
			children: make(map[string]*Stat),
		}
		parent.children[key] = s
	}

	return s
}

// AddTime comment
func (s *Stat) AddTime(startNanos int64) int64 {
	now := time.Now().UTC()

	endNanos := now.UnixNano()

	if endNanos < startNanos {
		log.Printf("%s: EndNanos is less than StartNanos", s.key)
		return 0
	}

	diff := endNanos - startNanos

	s.lastTime = now

	if s.count == 0 {
		s.firstTime = now
		s.firstNanos = diff
		s.minNanos = diff
		s.maxNanos = diff
	} else {
		if diff < s.minNanos {
			s.minNanos = diff
		}
		if diff > s.maxNanos {
			s.maxNanos = diff
		}
	}
	s.total += diff
	s.count++

	return endNanos
}

func (s *Stat) reset() {
	s.firstNanos = 0
	s.lastNanos = 0
	s.minNanos = 0
	s.maxNanos = 0
	s.total = 0
	s.count = 0
	s.firstTime = time.Time{}
	s.lastTime = time.Time{}

	for _, item := range s.children {
		item.reset()
	}
}

func (s *Stat) getRoot() *Stat {
	return rootItem
}

func (s *Stat) currentTimeNanos() int64 {
	return time.Now().UnixNano()
}

// Average comment
func (s *Stat) Average() int64 {
	if s.count == 0 {
		return 0
	}

	return s.total / s.count
}

func (s *Stat) printStatisticsHTML(p io.Writer, keysParam string) {
	fmt.Fprintf(p, "<html><head>\r\n")
	fmt.Fprintf(p, "<title>\r\n")
	fmt.Fprintf(p, "Maestro Statistics\r\n")
	fmt.Fprintf(p, "</title>\r\n")
	fmt.Fprintf(p, "<script type='text/javascript' src='/js/jquery-1.3.2.js'></script>")
	fmt.Fprintf(p, "<script type='text/javascript' src='/js/jquery.tablesorter.js'></script>")
	fmt.Fprintf(p, "<script type='text/javascript' src='/js/chili-1.8b.js'></script>")
	fmt.Fprintf(p, "<link rel='stylesheet' href='/css/statistics.css' type='text/css' media='print, projection, screen' />")
	fmt.Fprintf(p, "<script type='text/javascript'>\r\n")

	fmt.Fprintf(p, "$(document).ready(function() \r\n")
	fmt.Fprintf(p, "{ \r\n")
	fmt.Fprintf(p, "$('#myTable').tablesorter( {\r\n")
	fmt.Fprintf(p, "sortList: [[1,1]],\r\n")
	fmt.Fprintf(p, "debug: false,\r\n")
	fmt.Fprintf(p, "widgets: ['zebra'],\r\n")
	fmt.Fprintf(p, "headers: { 1: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "2: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "3: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "4: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "5: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "6: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "7: {sorter: 'currency'},\r\n")
	fmt.Fprintf(p, "8: {sorter: 'usLongDate'}\r\n")

	fmt.Fprintf(p, "}\r\n")
	fmt.Fprintf(p, "} )\r\n")

	fmt.Fprintf(p, "} \r\n")
	fmt.Fprintf(p, ")  \r\n")
	fmt.Fprintf(p, "</script>\r\n")
	fmt.Fprintf(p, "</head>\r\n")
	fmt.Fprintf(p, "<body>\r\n")

	// 		// 10/03/14 (SRL) - Adding of a reset button
	fmt.Fprintf(p, "<table width='100%'>\r\n")
	fmt.Fprintf(p, "<tr>\r\n")
	// 		// Title
	fmt.Fprintf(p, "<td style='vertical-align:middle;width:50%'>\r\n")
	fmt.Fprintf(p, "<h1>\r\n")
	fmt.Fprintf(p, "Maestro Statistics\r\n")
	fmt.Fprintf(p, "</h1>\r\n")
	fmt.Fprintf(p, "</td>\r\n")
	// 		// New button
	fmt.Fprintf(p, "<td align='right' style='vertical-align:middle;width:50%' >\r\n")
	// fmt.Fprintf(p, "<form border='0' cellpadding='0'>\r\n")
	// 		// Using location.replace here so that the history buffer is not messed up for going back a page.
	// fmt.Fprintf(p, "<input type='button' value='  Reset Statistics  ' onClick='window.location.replace(\"resetstatistics.html?%s\")'>\r\n", keysParam)
	// fmt.Fprintf(p, "</form>\r\n")
	fmt.Fprintf(p, "</td>\r\n")
	fmt.Fprintf(p, "</tr>\r\n")
	fmt.Fprintf(p, "</table>\r\n")
	// 		// End of change to add a reset button

	fmt.Fprintf(p, "<table id='myTable' class='tablesorter' border='0' cellpadding='0' cellspacing='1'>\r\n")
	fmt.Fprintf(p, "<thead>\r\n")
	fmt.Fprintf(p, "<tr>\r\n")
	fmt.Fprintf(p, "<th>Item</th>\r\n")
	fmt.Fprintf(p, "<th>total</th>\r\n")
	fmt.Fprintf(p, "<th>count</th>\r\n")
	fmt.Fprintf(p, "<th>first</th>\r\n")
	fmt.Fprintf(p, "<th>last</th>\r\n")
	fmt.Fprintf(p, "<th>min</th>\r\n")
	fmt.Fprintf(p, "<th>max</th>\r\n")
	fmt.Fprintf(p, "<th>average</th>\r\n")
	fmt.Fprintf(p, "<th>first run</th>\r\n")
	fmt.Fprintf(p, "<th>last run</th>\r\n")
	fmt.Fprintf(p, "</tr>\r\n")
	fmt.Fprintf(p, "</thead>\r\n")

	fmt.Fprintf(p, "<tbody>\r\n")

	now := time.Now().UTC()

	fmt.Fprintf(p, "<h2>Server started: %s [%s ago]</h2>\r\n", initTime.Format("2006-01-02 03:04:05.000"), HumanTime(time.Since(initTime)))

	for key, item := range s.getRoot().children {
		fmt.Fprintf(p, "<tr>\r\n")
		if len(item.children) > 0 {
			context := ""
			keysParam := ""
			fmt.Fprintf(p, "<td><a href='%s?%s' /></td>\r\n", context, keysParam)
		} else {
			fmt.Fprintf(p, "<td>%s</td>\r\n", key)
		}

		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.total)
		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.count)
		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.firstNanos)
		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.lastNanos)
		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.minNanos)
		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.maxNanos)
		fmt.Fprintf(p, "<td align='right'>%d</td>\r\n", item.Average())
		fmt.Fprintf(p, "<td align='right'>%s</td>\r\n", item.firstTime.Format("2006-01-02 03:04:05.000"))
		fmt.Fprintf(p, "<td align='right'>%s</td>\r\n", item.lastTime.Format("2006-01-02 03:04:05.000"))
		fmt.Fprintf(p, "</tr>\r\n")
	}

	// 		for (StatisticalItem item : statisticalItem.getChildren()) {
	// 				if (item != statisticalItem) {

	//
	// 		}

	fmt.Fprintf(p, "</tbody>\r\n")

	fmt.Fprintf(p, "</table>\r\n")
	fmt.Fprintf(p, "<p>Report time: %s</p>\r\n", now.Format("2006-01-02 03:04:05.000"))
	fmt.Fprintf(p, "<div align='right'>")
	// fmt.Fprintf(p, "<form>\r\n\r\n")
	// fmt.Fprintf(p, "<input type='button' value='  Back  ' onClick='history.go(-1)'>\r\n")
	// fmt.Fprintf(p, "</form>\r\n")
	fmt.Fprintf(p, "</div>\r\n")
	fmt.Fprintf(p, "</body></html>\r\n")

}

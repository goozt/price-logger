package parser

import (
	"dilogger/internal/model"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// The Parse function reads HTML content from a given URL, parses it to extract table rows, and sends each row to a channel for further processing.
func Parse(ch chan model.Product, wg *sync.WaitGroup, url string) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body := resp.Body
	if err != nil {
		log.Fatal(err)
	}
	doc, err := html.Parse(body)
	body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var tbody *html.Node
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.Tbody {
			tbody = n
		}
	}
	for n := range tbody.ChildNodes() {
		if n.Data == "tr" {
			ch <- ParseRow(n)
		}
	}
}

// The ParseRow function extracts product information from an HTML row and returns a model.Product struct.
func ParseRow(row *html.Node) model.Product {
	var name string
	var stock int
	var price float64
	var tds []*html.Node
	for td := range row.ChildNodes() {
		if td.Data == "td" {
			tds = append(tds, td)
		}
	}
	nameNode := tds[1]
	priceNode := tds[2]

	var nslist []string
	for d := range nameNode.Descendants() {
		if d.Type == html.TextNode && len(strings.TrimSpace(d.Data)) > 0 {
			nslist = append(nslist, strings.TrimSpace(d.Data))
		}
	}
	for d := range priceNode.Descendants() {
		if d.Type == html.ElementNode && d.DataAtom == atom.Ins {
			for e := range d.Descendants() {
				data := strings.TrimSpace(e.Data)
				if e.Type == html.TextNode && len(data) > 0 && !strings.Contains(data, "₹") {
					price, _ = strconv.ParseFloat(strings.Replace(data, ",", "", -1), 64)
				}
			}
		}
	}

	name = nslist[0]
	stock, _ = strconv.Atoi(strings.Replace(nslist[1], " in stock", "", 1))
	return model.Product{
		Name:      name,
		Stock:     int32(stock),
		Price:     float64(price),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

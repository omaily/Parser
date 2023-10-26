package shikimori

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
)

type MangaCopy struct {
	Url     string
	TitleEn string
	TitleRu string
	Date    string
}

func mainCopy() {

	fName := "data.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	c := colly.NewCollector()
	c.SetRequestTimeout(120 * time.Second)
	mangas := make([]Manga, 0)

	var pagesToScrape []string
	pageToScrape := "https://shikimori.one/mangas/?score=8"
	pagesDiscovered := []string{pageToScrape}

	c.OnHTML("div.pagination", func(e *colly.HTMLElement) {
		linkNext := e.ChildAttr("a.link-next", "href")
		if !containsCopy(pagesToScrape, linkNext) {
			if !containsCopy(pagesDiscovered, linkNext) {
				pagesToScrape = append(pagesToScrape, linkNext)
			}
			pagesDiscovered = append(pagesDiscovered, linkNext)
		}
	})

	c.OnHTML("section.l-content", func(l *colly.HTMLElement) {
		l.ForEach("div.cc-entries", func(i int, h *colly.HTMLElement) {
			chaild := l.ChildAttrs("article", "id")
			h.ForEach("a.cover", func(i int, h *colly.HTMLElement) {
				manga := Manga{}
				manga.link = h.Attr("href")
				manga.TitleEn = h.ChildText(".name-en")
				manga.TitleRu = h.ChildText(".name-ru")
				manga.Date = h.ChildText(".right")
				mangas = append(mangas, manga)
				writer.Write([]string{manga.link, manga.TitleEn, manga.TitleRu, manga.Date})
			})
			fmt.Println(chaild)
		})
	})

	c.OnScraped(func(response *colly.Response) {
		if len(pagesToScrape) != 0 {
			pageToScrape = pagesToScrape[0]
			pagesToScrape = pagesToScrape[1:]
			c.Visit(pageToScrape)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Got this error:", e)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		js, err := json.MarshalIndent(mangas, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile("manga.json", js, 0664); err != nil {
			fmt.Println("file knock on coffin")
		}
	})

	c.Visit(pageToScrape)
}

func containsCopy(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

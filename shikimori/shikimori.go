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

type Manga struct {
	link       string
	Type       string
	TitleEn    string
	TitleRu    string
	Date       string
	Rating     string
	Status     string
	Toms       string
	Part       string
	AuthorEn   string
	AuthorRu   string
	AuthorLink string
}

func Parse() {
	fName := "shikimori/data.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	c := colly.NewCollector(
		colly.CacheDir("./manga_cache"),
	)
	c.SetRequestTimeout(600 * time.Second)
	detailCollector := c.Clone()

	PageScrape := "https://shikimori.one/mangas"
	pageCounter := 40
	mangas := make([]*Manga, 0)
	manga := &Manga{}

	c.OnHTML("div.pagination", func(e *colly.HTMLElement) {
		PageScrape = e.ChildAttr("a.link-next", "href")
	})

	c.OnHTML("section.l-content", func(l *colly.HTMLElement) {
		l.ForEach("div.cc-entries", func(i int, h *colly.HTMLElement) {
			chaild := l.ChildAttrs("article", "id")
			h.ForEach("article", func(i int, h *colly.HTMLElement) {
				link := h.ChildAttr("a", "href")
				manga = &Manga{}
				manga.link = h.ChildAttr("a", "href")
				manga.TitleEn = h.ChildText(":nth-child(2) > .name-en")
				manga.TitleRu = h.ChildText(":nth-child(2) > .name-ru")
				manga.Date = h.ChildText(".right")
				detailCollector.Visit(link)
			})

			fmt.Println(PageScrape, "--", chaild)
		})
	})

	detailCollector.OnHTML(`div.l-content:first-child`, func(e *colly.HTMLElement) { //block
		log.Println("Manga found", e.Request.URL)

		e.ForEach("div.c-info-left > div.block > div > .line-container", func(i int, h *colly.HTMLElement) {
			key := h.ChildText(".key")
			switch key {
			case "Тип:":
				manga.Type = h.ChildText(".value")
			case "Тома:":
				manga.Toms = h.ChildText(".value")
			case "Главы:":
				manga.Part = h.ChildText(".value")
			case "Статус:":
				manga.Status = h.ChildAttr(".b-anime_status_tag", "data-text")
			}
		})
		manga.Rating = e.ChildText("div.b-rate > div.text-score > div.score-value")

		mangas = append(mangas, manga)
		writer.Write([]string{manga.link, manga.TitleEn, manga.TitleRu, manga.Rating, manga.Date, manga.Type, manga.Toms, manga.Part, manga.Status})
	})

	c.OnScraped(func(response *colly.Response) {
		if len(PageScrape) != 0 && pageCounter > 0 {
			pageCounter--
			c.Visit(PageScrape)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		js, err := json.MarshalIndent(mangas, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile("shikimori/manga.json", js, 0664); err != nil {
			fmt.Println("file knock on coffin")
		}
	})

	c.Visit(PageScrape)
}

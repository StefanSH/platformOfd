package platformOfd

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"strconv"
	"strings"
	"time"
)

type platformOfd struct {
	Username string
	Password string
	CSRF     string
}

type Receipt struct {
	ID       int
	FP       string
	Date     string
	Products []Product
	Link     string
	Price    int
	VatPrice int
}

type Product struct {
	Name     string
	Quantity int
	Price    int
	Vat      int
	VatPrice int
}

func PlatformOfd(Username string, Password string) *platformOfd {
	return &platformOfd{
		Username: Username,
		Password: Password,
	}
}

func (pf *platformOfd) GetReceipts(date time.Time) (receipts []Receipt, err error) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c.OnHTML("#login_form_id > div > input[type=hidden]", func(e *colly.HTMLElement) {
		csrf := e.Attr("value")
		err := c.Post("https://lk.platformaofd.ru/web/j_spring_security_check",
			map[string]string{
				"j_username": pf.Username,
				"j_password": pf.Password,
				"_csrf":      csrf,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		startDate := time.Time{}
		endDate := time.Time{}
		for h := 1; h <= date.Hour(); h++ {
			startDate = time.Date(date.Year(), date.Month(), date.Day(), h-1, 0, 0, 0, date.Location())
			endDate = time.Date(date.Year(), date.Month(), date.Day(), h, 0, 0, 0, date.Location())
			receipts, err = pf.getChecksLink(c.Clone(), startDate, endDate)
		}
	})

	err = c.Visit("https://lk.platformaofd.ru/web/login")
	if err != nil {
		return receipts, err
	}

	return receipts, nil
}

func (pf *platformOfd) getChecksLink(c *colly.Collector, startDate time.Time, endDate time.Time) (receipts []Receipt, err error) {
	c.OnHTML("#cheques-search-content > div > div > div > table > tbody > tr", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if link == "/web/auth/cheques/reports" {
			return
		}
		pLinkSource := strings.Split(link, "?")
		pLink := strings.Split(pLinkSource[0], "/")
		//https://lk.platformaofd.ru/web/auth/cheques/details/<id>/<date>/<fp>?date=28.11.2019+17%3A42
		//https://lk.platformaofd.ru/web/noauth/cheque/id?id=<id>&date=<date>&fp=<fp>
		products, _ := pf.getCheck(c.Clone(), fmt.Sprintf("/web/noauth/cheque/id?id=%s&date=%s&fp=%s", pLink[5], pLink[6], pLink[7]))
		id, err := strconv.Atoi(pLink[5])
		if err != nil {
			log.Printf("%v", err)
		}
		receipt := Receipt{
			ID:       id,
			FP:       pLink[7],
			Date:     pLink[6],
			Products: products,
			Link:     link,
			Price:    0,
			VatPrice: 0,
		}
		receipts = append(receipts, receipt)
	})
	//https://lk.platformaofd.ru/web/auth/cheques?start=27.11.2019+13%3A00&end=27.11.2019+13%3A00
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru/web/auth/cheques?start=%s&end=%s", startDate.Format("02.01.2006+15%3A04"), endDate.Format("02.01.2006+15%3A04")))
	if err != nil {
		return receipts, err
	}

	return receipts, nil
}

func (pf *platformOfd) getCheck(c *colly.Collector, link string) (product []Product, err error) {
	c.OnHTML("div.check-product-name", func(e *colly.HTMLElement) {
		productName := e.Text
		pr := Product{
			Name:     productName,
			Quantity: 0,
			Price:    0,
			Vat:      0,
			VatPrice: 0,
		}
		product = append(product, pr)
	})

	link = strings.Replace(link, ":", "%3A", -1)
	link = strings.Replace(link, " ", "%20", -1)
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru%s", link))
	if err != nil {
		return product, err
	}

	return product, nil
}

package platformOfd

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"strings"
	"time"
)

type platformOfd struct {
	Username string
	Password string
	CSRF     string
}

type Receipt struct {
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

			fmt.Printf("%s - %s", startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
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
		//ch := d.Clone()
		link := e.Attr("href")
		log.Printf("Link to href: %s", link)
		pLink := strings.Split(link, "/")
		//https://lk.platformaofd.ru/web/auth/cheques/details/<id>/<date>/<fp>?date=28.11.2019+17%3A42
		//https://lk.platformaofd.ru/web/noauth/cheque/id?id=<id>&date=<date>&fp=<fp>
		products, _ := pf.getCheck(c.Clone(), fmt.Sprintf("/web/noauth/cheque/id?id=%s&date=%s&fp=%s", pLink[4], pLink[5], pLink[6]))

		receipt := Receipt{
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
		log.Printf("ProductName: %s", productName)
	})

	link = strings.Replace(link, ":", "%3A", -1)
	link = strings.Replace(link, " ", "%20", -1)
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru%s", link))
	if err != nil {
		return product, err
	}

	return product, nil
}

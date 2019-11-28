package platformOfd

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"strings"
)

type platformOfd struct {
	Username string
	Password string
	CSRF     string
}

type Receipt struct {
	Products []Product
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

func (pf *platformOfd) GetReceipts(startDate string, endDate string) (receipts []Receipt, err error) {
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
		receipts, err = pf.getChecksLink(c.Clone(), startDate, endDate)
	})

	err = c.Visit("https://lk.platformaofd.ru/web/login")
	if err != nil {
		return receipts, err
	}

	return receipts, nil
}

func (pf *platformOfd) getChecksLink(c *colly.Collector, startDate string, endDate string) (receipts []Receipt, err error) {
	c.OnHTML("#cheques-search-content > div > div > div > table > tbody > tr", func(e *colly.HTMLElement) {
		//ch := d.Clone()
		link := e.Attr("href")
		log.Printf("Link to href: %s", link)
		receipt, _ := pf.getCheck(c.Clone(), link)
		receipts = append(receipts, receipt)
		/*ch.OnHTML("a.btn.btn-default.text-nowrap", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			log.Printf("Link to check: %s", link)

		})
		link = strings.Replace(link, ":", "%3A", -1)
		link = strings.Replace(link, " ", "%20", -1)
		ch.Visit(fmt.Sprintf("https://lk.platformaofd.ru%s", link))*/
	})
	err = c.Visit("https://lk.platformaofd.ru/web/auth/cheques?start=27.11.2019+13%3A00&end=27.11.2019+13%3A00")
	if err != nil {
		return receipts, err
	}

	return receipts, nil
}

func (pf *platformOfd) getCheck(c *colly.Collector, link string) (receipt Receipt, err error) {
	c.OnHTML("div.check-product-name", func(e *colly.HTMLElement) {
		productName := e.Text
		log.Printf("ProductName: %s", productName)
	})

	link = strings.Replace(link, ":", "%3A", -1)
	link = strings.Replace(link, " ", "%20", -1)
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru%s", link))
	if err != nil {
		return receipt, err
	}

	return receipt, nil
}

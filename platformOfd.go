package platformOfd

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"net/url"
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
	FD       string
	Date     string
	Products []Product
	Link     string
	Price    int
	VatPrice int
}

type Product struct {
	Name       string
	Quantity   int
	Price      int
	Vat        int
	VatPrice   int
	TotalPrice string
	FP         string
	FD         string
	FN         string
	Time       string
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
			receipt, _ := pf.getChecksLink(c.Clone(), startDate, endDate)
			receipts = append(receipts, receipt...)
		}
	})

	err = c.Visit("https://lk.platformaofd.ru/web/login")
	if err != nil {
		return receipts, err
	}

	return receipts, nil
}

func (pf *platformOfd) getChecksLink(c *colly.Collector, startDate time.Time, endDate time.Time) (receipts []Receipt, err error) {
	//log.Println(url.QueryEscape(fmt.Sprintf("https://lk.platformaofd.ru/web/auth/cheques?start=%s&end=%s", startDate.Format("02.01.2006 15:04"), endDate.Format("02.01.2006 15:04"))))
	c.OnHTML("#cheques-search-content > div > div > div > table > tbody > tr", func(e *colly.HTMLElement) {
		check := c.Clone()
		link := e.Attr("href")

		if link == "/web/auth/cheques/reports" {
			return
		}
		pLinkSource := strings.Split(link, "?")
		pLink := strings.Split(pLinkSource[0], "/")
		cLink := fmt.Sprintf("/web/noauth/cheque/id?id=%s&date=%s&fp=%s", pLink[5], pLink[6], pLink[7])
		checkNumber, products, _ := pf.getCheck(check, cLink)

		if len(products) > 0 {
			totalPrice, err := strconv.ParseFloat(products[0].TotalPrice, 64)
			if err != nil {
				log.Printf("%v", err)
			}
			receipt := Receipt{
				ID:       checkNumber,
				FP:       pLink[7],
				FD:       products[0].FD,
				Date:     pLink[6],
				Products: products,
				Link:     fmt.Sprintf("https://lk.platformaofd.ru%s", cLink),
				Price:    int(totalPrice * float64(100)),
				VatPrice: 0,
			}
			receipts = append(receipts, receipt)
		}
	})
	//https://lk.platformaofd.ru/web/auth/cheques?start=27.11.2019+13%3A00&end=27.11.2019+13%3A00
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru/web/auth/cheques?start=%s&end=%s", url.QueryEscape(startDate.Format("02.01.2006 15:04")), url.QueryEscape(endDate.Format("02.01.2006 15:04"))))
	if err != nil {
		return receipts, err
	}

	return receipts, nil
}

func (pf *platformOfd) getFd(c *colly.Collector, link string) (fd string, err error) {
	c.OnHTML("div.check-product-name", func(e *colly.HTMLElement) {
		fd = e.Attr("src")
	})
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru%s", link))
	if err != nil {
		return "", err
	}
	return fd, nil
}

func (pf *platformOfd) getCheck(c *colly.Collector, link string) (checkNumber int, product []Product, err error) {
	c.OnHTML("h1.check-headline>span", func(e *colly.HTMLElement) {
		checkNumber, err = strconv.Atoi(e.Text)
		if err != nil {
			log.Printf("%v", err)
		}
	})

	c.OnHTML("div", func(e *colly.HTMLElement) {
		pr := Product{}
		e.ForEach("div.check-product-name", func(i int, e *colly.HTMLElement) {
			productName := e.Text
			pr.Name = productName
		})
		e.ForEach("div.check-qr>img", func(i int, e *colly.HTMLElement) {
			u, err := url.Parse(e.Attr("src"))
			if err != nil {
				log.Println(err)
			}

			query := u.Query()
			pr.FP = query["fp"][0]
			pr.FN = query["fn"][0]
			pr.FD = query["i"][0]
			pr.TotalPrice = query["s"][0]
			pr.Time = query["t"][0]
		})
		if pr.Name != "" {
			product = append(product, pr)
		}
	})

	link = strings.Replace(link, ":", "%3A", -1)
	link = strings.Replace(link, " ", "%20", -1)
	err = c.Visit(fmt.Sprintf("https://lk.platformaofd.ru%s", link))
	if err != nil {
		return checkNumber, product, err
	}

	return checkNumber, product, nil
}

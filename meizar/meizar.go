package meizar

import (
	"../rule"
	"../store"
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	mysql_host = "10.173.32.9"
	mysql_user = "taskdone"
	mysql_pwd  = "wangfei808"
	mysql_db   = "meizi"
)

func New(dir string, startPage int, r rule.Rule, cookie string, client *http.Client, pageSort int, s store.Store) *Meizar {
	return &Meizar{dir: dir, currentPage: startPage, userCookie: cookie, r: r, client: client, pageSort: pageSort, s: s}
}

type Meizar struct {
	dir         string
	currentPage int
	userCookie  string
	client      *http.Client
	r           rule.Rule
	pageSort    int
	s           store.Store
}

func (p *Meizar) Start() {
	if !p.isExist(p.dir) {
		if err := os.Mkdir(p.dir, 0777); err != nil {
			panic("can not mkdir " + p.dir)
		}
	}

	db, err := sql.Open("mysql", mysql_user+":"+mysql_pwd+"@tcp("+mysql_host+"):3306/"+mysql_db+"?charset=utf8")
	if err != nil {
		panic("mysql error " + err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic("mysql ping error " + err)
	}

	for p.currentPage > 0 {
		time.Sleep(1e9)
		p.parsePage(p.r.UrlRule() + p.r.PageRule(p.currentPage))
		if p.pageSort == 1 {
			p.currentPage++
		} else {
			p.currentPage--
		}
	}
}

func (p *Meizar) parsePage(db *sql.DB, url string) {
	req := p.buildRequest(url)
	resp, err := p.client.Do(req)

	if err != nil {
		fmt.Println("failed parse " + url)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(url + "-->" + strconv.Itoa(resp.StatusCode))
		return
	}

	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	img, err := p.parseImageUrl(bytes.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, item := range img {
		go p.downloadImage(db, item)
		//go p.uploadImage(item)
	}
}

func (p *Meizar) buildRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.152 Safari/537.36")
	req.Header.Set("Cookie", p.userCookie)
	return req
}

func (p *Meizar) parseImageUrl(reader io.Reader) (res []string, err error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	p.r.ImageRule(doc, func(image string) {
		res = append(res, image)
	})

	return res, nil
}

func (p *Meizar) downloadImage(db *sql.DB, url string) {
	fileName := p.getNameFromUrl(url)
	if p.isExist(p.dir + fileName) {
		fmt.Println("already download " + fileName)
		return
	}

	req := p.buildRequest(url)
	resp, err := p.client.Do(req)
	if err != nil {
		fmt.Println("failed download " + url)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("failed download " + url)
		return
	}

	defer func() {
		resp.Body.Close()
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	//fmt.Println("begin download " + fileName)
	localFile, _ := os.OpenFile(p.dir+fileName, os.O_CREATE|os.O_RDWR, 0777)
	if _, err := io.Copy(localFile, resp.Body); err != nil {
		panic("failed save " + fileName)
	}

	p.s.Upload(localFile.Name())
	p.addRecord(db, fileName)
	//os.Remove(p.dir + fileName)

	//fmt.Println("success download " + fileName)
}

func (p *Meizar) uploadImage(url string) {
	p.s.Upload(url)
}

func (p *Meizar) isExist(dir string) bool {
	_, err := os.Stat(dir)
	if err == nil {
		return true
	}

	return os.IsExist(err)
}

func (p *Meizar) getNameFromUrl(url string) string {
	arr := strings.Split(url, "/")
	return arr[len(arr)-1]
}

func (p *Meizar) addRecord(db *sql.DB, url string) {
	stmt, err = db.Prepare("insert into meizi.imgs(url) values(?)")
	if err != nil {
		fmt.Println(err)
	}
	defer stmp.Close()
	stmp.Exec(url)
}

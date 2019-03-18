package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/axgle/mahonia"
)

//编码转换
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// 获取整页数据
func HttpGet(url string) (result string, err error) {
	resp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()
	var html string
	//循环爬取整页数据
	buf := make([]byte, 4096)
	for {
		n, err2 := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
		}
		html += string(buf[:n])
	}

	result = ConvertToString(html, "gbk", "utf-8")
	//fmt.Println("页面内容")
	//fmt.Println(result)
	return
}

// 解析数据
func SpiderPage(page int, pipe chan int) {
	//拼接URL地址
	url := "http://bang.dangdang.com/books/newhotsales/01.00.00.00.00.00-recent7-0-0-1-" + strconv.Itoa(page)
	//
	result, err := HttpGet(url)
	if err != nil {
		fmt.Println("爬取 " + strconv.Itoa(page) + " 页 发生错误!")
		return
	}
	//fmt.Println("结果:" + result)

	nameSlice := make([]string, 0)   //保存名称slice
	priceSlice := make([]string, 0)  //保存价格slice
	dateSlice := make([]string, 0)   //保存出版日期slice
	authorSlice := make([]string, 0) //保存作者slice
	unitSlice := make([]string, 0)   //保存出版社slice

	//链接地址
	rest0 := regexp.MustCompile(`<div class="name"><a href="(?s:(.*?))"`)
	urls := rest0.FindAllStringSubmatch(result, -1) //-1全部 1第一条
	for _, url := range urls {
		//fmt.Println("链接地址:", url[1])
		name, price, date, author, unit, err := SpiderBookPage(url[1])
		if err != nil {
			continue
		}
		nameSlice = append(nameSlice, name)
		priceSlice = append(priceSlice, price)
		dateSlice = append(dateSlice, date)
		authorSlice = append(authorSlice, author)
		unitSlice = append(unitSlice, unit)
		/*
			fmt.Println("-------------------------")
			fmt.Println("书名:", name)
			fmt.Println("价格:", price)
			fmt.Println("出版时间:", date)
			fmt.Println("作者:", author)
			fmt.Println("出版社:", unit)*/
	}

	//保存到文件中
	save2File(page, nameSlice, priceSlice, dateSlice, authorSlice, unitSlice)

	pipe <- page
}

//解析页面上的 书名 价格 出版时间 作者 出版社
func SpiderBookPage(url string) (name, price, date, author, unit string, err error) {
	result, err0 := HttpGet(url)
	if err0 != nil {
		err = err0
		return
	}
	//fmt.Println("结果:" + result)

	//书名地址
	rest0 := regexp.MustCompile(`<h1 title="(?s:(.*?))">`)
	names := rest0.FindAllStringSubmatch(result, -1) //-1全部 1第一条
	for _, nameTmp := range names {
		name = nameTmp[1]
	}

	//价格
	rest1 := regexp.MustCompile(`&yen;</span>(?s:(.*?)) `)
	prices := rest1.FindAllStringSubmatch(result, 2) //-1全部 2第二条
	for _, priceTmp := range prices {
		price = priceTmp[1]
	}

	//出版时间
	rest2 := regexp.MustCompile(`<span class="t1">出版时间:(?s:(.*?))&nbsp;</span>`)
	dates := rest2.FindAllStringSubmatch(result, -1) //-1全部 1第一条
	for _, dateTmp := range dates {
		date = dateTmp[1]
	}

	//作者
	rest3 := regexp.MustCompile(`target="_blank" dd_name="作者">(?s:(.*?))</a>`)
	authors := rest3.FindAllStringSubmatch(result, 1) //-1全部 1第一条
	for _, authorTmp := range authors {
		author = authorTmp[1]
	}

	//出版社
	rest4 := regexp.MustCompile(`dd_name="出版社">(?s:(.*?))</a></span>`)
	units := rest4.FindAllStringSubmatch(result, -1) //-1全部 1第一条
	for _, unitTmp := range units {
		unit = unitTmp[1]
	}

	return
}

//保存到文件中
func save2File(page int, nameSlice, priceSlice, dateSlice, authorSlice, unitSlice []string) {
	f, err := os.Create("page" + strconv.Itoa(page) + ".txt")
	if err != nil {
		fmt.Println("创建文件错误!")
	}
	defer f.Close()

	count := len(nameSlice)
	for i := 0; i < count; i++ {
		/*
			fmt.Println("-------------------------")
			fmt.Println("书名:", bookNames[i][1])
			fmt.Println("价格:", prices[i][1])
			fmt.Println("推荐:", tuijians[i][1])
			fmt.Println("出版时间:", publishs[i][1])
			fmt.Println("作者:", authors[i][1])
		*/
		f.WriteString("-------------------------" + "\n")
		f.WriteString("书名:" + nameSlice[i] + "\n")
		f.WriteString("价格:" + priceSlice[i] + "\n")
		f.WriteString("出版时间:" + dateSlice[i] + "\n")
		f.WriteString("作者:" + authorSlice[i] + "\n")
		f.WriteString("出版社:" + unitSlice[i] + "\n")
	}

}

//主工作方法
func Work(start int, end int) {
	pipe := make(chan int)
	for i := start; i <= end; i++ {
		fmt.Printf("正在爬取 %d 页 \n", i)
		go SpiderPage(i, pipe)
	}

	for i := start; i <= end; i++ {
		fmt.Printf("爬取 %d 页结束 \n", <-pipe)
	}

}

func main() {
	var beginPage int
	var finishPage int
	fmt.Println("网络爬虫启动...")
	//fmt.Println("请输入开始页数:")
	//fmt.Scan(&beginPage)
	//fmt.Println("请输入结束页数:")
	//fmt.Scan(&finishPage)
	beginPage = 1
	finishPage = 5
	Work(beginPage, finishPage)
}

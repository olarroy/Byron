package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/ttacon/chalk"
)

type Article struct {
	Id          string
	Url         string
	DownloadUrl string
	Title       string
	Isbn        string
	Year        string
	Publisher   string
	Author      string
	Extension   string
	Page        string
	Language    string
	Size        string
	Time        string
}

func LIBGENDownloadAll(search string) {

	/*
		Libgen.is
		with a given search, get all the urls from where to download the books
	*/

	var (
		r       = regexp.MustCompile("<a href=.book/index.php.md5=([^\"']*)")
		AllUrls = []string{}
		count   = 0
	)
	for i := 1; i < 120000; i++ {

		//resp, err := http.Get("https://libgen.is/search.php?&res=100&req=" + search + "&phrase=1&view=simple&column=def&sort=def&sortmode=ASC&page=" + strconv.Itoa(i))
		resp, err := http.Get("https://libgen.is/search.php?mode=last&view=simple&phrase=0&timefirst=&timelast=&sort=def&sortmode=ASC&page=80" + strconv.Itoa(i))

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		html, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		htmlFormat := string(html)

		matches := r.FindAllStringSubmatch(htmlFormat, -1)

		fmt.Println(chalk.Green.Color("Processing page " + strconv.Itoa(i)))

		if len(matches) > 1 {
			for _, m := range matches {
				fmt.Println(chalk.Green.Color("Saving " + m[1]))
				AllUrls = append(AllUrls, "https://libgen.is/book/index.php?md5="+m[1])
				count++
			}
		} else {
			break
		}

	}

	log.Println("Total Articles:", count)

	ProcessUrls(AllUrls, search)
}

func ProcessUrls(AllUrls []string, search string) {

	for _, u := range AllUrls {

		/*
			For each Url
		*/
		resp, err := http.Get(u)

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		articleHtml, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		articleHtmlFormat := string(articleHtml)

		var (
			//All regex to get the data from
			ArticleTitle     = regexp.MustCompile("<title>Library Genesis:([^<]*)")
			ArticleAuthors   = regexp.MustCompile("Author.s.:</font></nobr></td><td colspan=3><b>([^<]*)")
			ArticlePublisher = regexp.MustCompile("Publisher:</font></nobr></td><td>([^<]*)")
			ArticleYear      = regexp.MustCompile("Year:</font></nobr></td><td>([^<]*)")
			ArticleLang      = regexp.MustCompile("Language:</font></nobr></td><td>([^<]*)")
			ArticleIsbn      = regexp.MustCompile("ISBN:</font></td><td>([^<]*)")
			ArticleTime      = regexp.MustCompile("Time modified:</font></nobr></td><td>([^<]*)")
			ArticleSize      = regexp.MustCompile("Size:</font></nobr></td><td>([^<]*)")
			ArticlePages     = regexp.MustCompile("Pages .biblio.tech.:</font></nobr></td><td>([^<]*)")
			ArticleId        = regexp.MustCompile("ID:</font></nobr></td><td>([^<]*)")
			ArticleExtension = regexp.MustCompile("Extension:</font></nobr></td><td>([^<]*)")
			ArticleDownload  = regexp.MustCompile("align=.center.><a href=.([^\"']*). title=.Gen.lib.rus.ec.")
		)

		AllArticles := ReadArticles(search)

		active := 0
		for _, art := range AllArticles {

			if art.Id == ArticleId.FindStringSubmatch(articleHtmlFormat)[1] {
				active = 1
			}
		}

		if active != 1 {

			/*
				Append and download because it's new
			*/

			AllArticles = append(AllArticles, Article{
				Id:          ArticleId.FindStringSubmatch(articleHtmlFormat)[1],
				Url:         u,
				DownloadUrl: ArticleDownload.FindStringSubmatch(articleHtmlFormat)[1],
				Title:       ArticleTitle.FindStringSubmatch(articleHtmlFormat)[1],
				Isbn:        ArticleIsbn.FindStringSubmatch(articleHtmlFormat)[1],
				Year:        ArticleYear.FindStringSubmatch(articleHtmlFormat)[1],
				Publisher:   ArticlePublisher.FindStringSubmatch(articleHtmlFormat)[1],
				Author:      ArticleAuthors.FindStringSubmatch(articleHtmlFormat)[1],
				Extension:   ArticleExtension.FindStringSubmatch(articleHtmlFormat)[1],
				Page:        ArticlePages.FindStringSubmatch(articleHtmlFormat)[1],
				Language:    ArticleLang.FindStringSubmatch(articleHtmlFormat)[1],
				Size:        ArticleSize.FindStringSubmatch(articleHtmlFormat)[1],
				Time:        ArticleTime.FindStringSubmatch(articleHtmlFormat)[1],
			})

			fmt.Println(chalk.Green.Color("Added correctly: " + ArticleTitle.FindStringSubmatch(articleHtmlFormat)[1]))
			WriteInFile(search, AllArticles)

			/*
				After that, download the file and save it
			*/
			FileDownload(
				ArticleDownload.FindStringSubmatch(articleHtmlFormat)[1],  //Download url
				ArticleId.FindStringSubmatch(articleHtmlFormat)[1],        //ID
				ArticleExtension.FindStringSubmatch(articleHtmlFormat)[1], //extension ex. .pdf
			)

		} else {
			fmt.Println(chalk.Magenta.Color("Already Exists in the register"))
		}

	}

	fmt.Println(chalk.Green.Color("All the documents were Downloaded :) "))
}

func FileDownload(URL, ID, format string) {

	/*
		Check html, regex the download link and save the file :)
	*/

	resp, err := http.Get(URL)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	DownloadHtml, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	DownloadHtmlFormat := string(DownloadHtml)
	downloadlinkRegex := regexp.MustCompile("<h2><a href=.([^\"']*)")
	downloadlink := downloadlinkRegex.FindStringSubmatch(DownloadHtmlFormat)[1]

	DownloadPDF(downloadlink, ID+"."+format)

}

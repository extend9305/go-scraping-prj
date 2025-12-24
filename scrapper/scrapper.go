package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	location string
	skills   []string
	date     string
}

func Scrape(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=" + term

	var result []extractedJob
	totalPages := getPages(baseURL)
	chnal := make(chan []extractedJob)

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, chnal)
	}

	for i := 0; i < totalPages; i++ {
		extractJobs := <-chnal
		result = append(result, extractJobs...)
	}

	writeJobs(result)
	fmt.Println("Done extracted:", len(result))

}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	headers := []string{"ID", "TITLE", "LOCATION", "SKILLS", "DATE"}
	wErr := csvWriter.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.location, strings.Join(job.skills, ","), job.date}
		wErr := csvWriter.Write(jobSlice)
		checkErr(wErr)
	}
}

func getPage(page int, baseUrl string, ch chan<- []extractedJob) {

	var jobs []extractedJob
	pageUrl := baseUrl + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println(pageUrl)
	resp, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(resp)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, s *goquery.Selection) {
		job := extractJobCard(s)
		jobs = append(jobs, job)
	})
	ch <- jobs
}

func extractJobCard(s *goquery.Selection) extractedJob {
	id := s.AttrOr("value", "")
	title, _ := s.Find("h2.job_tit a").Attr("title")
	location := ""
	s.Find("div.job_condition a").Each(func(i int, s *goquery.Selection) {
		location += s.Text() + " "
	})
	location = strings.Trim(location, " \n")
	skills := []string{}
	s.Find("div.job_sector a").Each(func(i int, s *goquery.Selection) {
		skills = append(skills, s.Text())
	})
	endDate := s.Find("div.job_date span.date").Text()
	return extractedJob{
		id:       id,
		title:    title,
		location: location,
		skills:   skills,
		date:     endDate,
	}
}

////saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=52603519

func getPages(baseURL string) int {
	currentPage := 1
	for {
		res, err := http.Get(baseURL + "&recruitPage=" + strconv.Itoa(currentPage))
		checkErr(err)
		checkCode(res)
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		checkErr(err)

		//다음 페이지 존재시 추가 로직

		nextPage, exists := doc.Find("div.pagination a.btnNext").Attr("page")
		if !exists {
			doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
				lastNum := s.Find("span").Length()
				currentPage += lastNum - 1 //마지막 뎁스 처리
			})
			break
		}

		num, err := strconv.Atoi(nextPage)
		checkErr(err)
		currentPage = num
	}

	return currentPage
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed", res.StatusCode)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

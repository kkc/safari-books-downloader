package safari

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	logrus "github.com/Sirupsen/logrus"
)

type ChapterMeta struct {
	Archive         string   `json:"archive"`
	Content         string   `json:"content"`
	URL             string   `json:"url"`
	NaturalKey      []string `json:"natural_key"`
	FullPath        string   `json:"full_path"`
	MinutesRequired float64  `json:"minutes_required"`
	NextChapter     struct {
		URL    string `json:"url"`
		Title  string `json:"title"`
		WebURL string `json:"web_url"`
	} `json:"next_chapter"`
	PreviousChapter struct {
		URL    string `json:"url"`
		Title  string `json:"title"`
		WebURL string `json:"web_url"`
	} `json:"previous_chapter"`
	Stylesheets []struct {
		FullPath    string `json:"full_path"`
		URL         string `json:"url"`
		OriginalURL string `json:"original_url"`
	} `json:"stylesheets"`
	Images               []interface{} `json:"images"`
	AssetBaseURL         string        `json:"asset_base_url"`
	WebURL               string        `json:"web_url"`
	LastPosition         interface{}   `json:"last_position"`
	Videoclips           []interface{} `json:"videoclips"`
	PublisherScripts     string        `json:"publisher_scripts"`
	PublisherScriptFiles []interface{} `json:"publisher_script_files"`
	AllowScripts         bool          `json:"allow_scripts"`
	Videoclip            interface{}   `json:"videoclip"`
	AcademicExcluded     bool          `json:"academic_excluded"`
	Subjects             []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"subjects"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Cover            string      `json:"cover"`
	BookTitle        string      `json:"book_title"`
	Updated          time.Time   `json:"updated"`
	SiteStyles       []string    `json:"site_styles"`
	CreatedTime      time.Time   `json:"created_time"`
	LastModifiedTime time.Time   `json:"last_modified_time"`
	Filename         string      `json:"filename"`
	Path             string      `json:"path"`
	EpubProperties   interface{} `json:"epub_properties"`
	HeadExtra        interface{} `json:"head_extra"`
	HasVideo         bool        `json:"has_video"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	VirtualPages     int         `json:"virtual_pages"`
}

type Meta struct {
	URL        string   `json:"url"`
	NaturalKey []string `json:"natural_key"`
	Authors    []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Subjects []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"subjects"`
	Topics []struct {
		Score          float64 `json:"score"`
		Name           string  `json:"name"`
		Slug           string  `json:"slug"`
		UUID           string  `json:"uuid"`
		EpubIdentifier string  `json:"epub_identifier"`
	} `json:"topics"`
	Publishers []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		Slug string `json:"slug"`
	} `json:"publishers"`
	Chapters        []string `json:"chapters"`
	Cover           string   `json:"cover"`
	ChapterList     string   `json:"chapter_list"`
	Toc             string   `json:"toc"`
	FlatToc         string   `json:"flat_toc"`
	WebURL          string   `json:"web_url"`
	LastChapterRead struct {
		URL    string `json:"url"`
		Title  string `json:"title"`
		WebURL string `json:"web_url"`
	} `json:"last_chapter_read"`
	AcademicExcluded        bool        `json:"academic_excluded"`
	OpfUniqueIdentifierType string      `json:"opf_unique_identifier_type"`
	CreatedTime             time.Time   `json:"created_time"`
	LastModifiedTime        time.Time   `json:"last_modified_time"`
	Identifier              string      `json:"identifier"`
	Name                    string      `json:"name"`
	Title                   string      `json:"title"`
	Format                  string      `json:"format"`
	ContentFormat           string      `json:"content_format"`
	Source                  string      `json:"source"`
	OrderableTitle          string      `json:"orderable_title"`
	HasStylesheets          bool        `json:"has_stylesheets"`
	Description             string      `json:"description"`
	Isbn                    string      `json:"isbn"`
	Issued                  string      `json:"issued"`
	Language                string      `json:"language"`
	Rights                  string      `json:"rights"`
	Updated                 time.Time   `json:"updated"`
	OrderableAuthor         string      `json:"orderable_author"`
	PurchaseLink            interface{} `json:"purchase_link"`
	PublisherResourceLinks  struct {
		Errata string `json:"Errata"`
	} `json:"publisher_resource_links"`
	IsFree          bool        `json:"is_free"`
	IsSystemBook    bool        `json:"is_system_book"`
	IsActive        bool        `json:"is_active"`
	IsHidden        bool        `json:"is_hidden"`
	VirtualPages    int         `json:"virtual_pages"`
	DurationSeconds interface{} `json:"duration_seconds"`
	Pagecount       int         `json:"pagecount"`
}

type Chapter struct {
	Filename       string
	Images         []string
	AssetBaseURL   string
	Title          string
	Content        string
	Id             string
	Order          int
	StylesheetsURL []string
}

type Book struct {
	id         string
	toc        map[string]TocContent
	chapters   map[int]Chapter
	stylesheet string
	meta       Meta
	sync.RWMutex
}

type jsonBook struct {
	Title       string
	Uuid        string
	Language    string
	Author      []string
	Cover       string
	Description string
	Publisher   []string
	Stylesheet  string
	Chapters    []Chapter
}

type Safari struct {
	baseUrl      string
	clientSecret string
	clientId     string
	books        map[string]Book
	accessToken  string
	sync.RWMutex
}

func NewSafari() *Safari {
	// clientSecret and clientId comes from https://github.com/nicohaenggi/SafariBooks-Downloader/blob/master/lib/safari/index.js
	safari := &Safari{
		baseUrl:      "https://www.safaribooksonline.com",
		clientSecret: "f52b3e30b68c1820adb08609c799cb6da1c29975",
		clientId:     "446a8a270214734f42a7",
		books:        make(map[string]Book),
	}

	return safari
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func (s *Safari) adjustOrderByChapterNumber(chapters map[int]Chapter) ([]Chapter, error) {
	total := len(chapters)
	var chapters_slice []Chapter

	for i := 0; i < total; i++ {
		chapters_slice = append(chapters_slice, chapters[i])
	}

	return chapters_slice, nil
}

// Get result by using book id
func (s *Safari) FetchBookById(id string, username string, password string) ([]byte, error) {
	// check input format

	err := s.authorizeUser(username, password)
	if err != nil {
		return nil, err
	}
	err = s.fetchMeta(id)
	if err != nil {
		return nil, err
	}
	_ = s.fetchTOC(id)
	err = s.fetchChapters(id)
	if err != nil {
		return nil, err
	}
	err = s.fetchStylesheet(id)
	if err != nil {
		return nil, err
	}

	var author []string
	for _, a := range s.books[id].meta.Authors {
		author = append(author, a.Name)
	}

	var publisher []string
	for _, p := range s.books[id].meta.Publishers {
		publisher = append(publisher, p.Name)
	}

	chapters, err := s.adjustOrderByChapterNumber(s.books[id].chapters)
	if err != nil {
		return nil, err
	}

	response := &jsonBook{
		Title:       s.books[id].meta.Title,
		Uuid:        s.books[id].meta.Identifier,
		Language:    s.books[id].meta.Language,
		Author:      author[:],
		Cover:       s.books[id].meta.Cover,
		Description: s.books[id].meta.Description,
		Publisher:   publisher[:],
		Stylesheet:  s.books[id].stylesheet,
		Chapters:    chapters,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type AuthResponse struct {
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// Login safari and get the access token
func (s *Safari) authorizeUser(username string, password string) error {
	// http request
	uri := s.baseUrl + "/oauth2/access_token/"
	form := url.Values{
		"client_id":     {s.clientId},
		"client_secret": {s.clientSecret},
		"grant_type":    {"password"},
		"username":      {username},
		"password":      {password},
	}
	resp, err := http.PostForm(uri, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Login fail, please double check your username and password")
	}

	logrus.Info("login successfully")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var stuff AuthResponse
	err = json.Unmarshal(body, &stuff)
	if err != nil {
		return err
	}

	s.accessToken = stuff.AccessToken
	return nil
}

// Fetch safari resources by given url
func (s *Safari) fetchResource(url string) (string, error) {
	uri := s.baseUrl + "/" + url

	logrus.Info("fetch uri " + uri + " with token " + s.accessToken)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("Error: status code != 200, actual status code '" + resp.Status + "'")
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return string(body), nil
}

func (s *Safari) fetchMeta(id string) error {
	url := "/api/v1/book/" + id
	body, err := s.fetchResource(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var meta Meta
	err = json.Unmarshal([]byte(body), &meta)
	if err != nil {
		fmt.Println(err)
		return err
	}
	s.books[id] = Book{
		id:         id,
		toc:        make(map[string]TocContent),
		chapters:   make(map[int]Chapter),
		stylesheet: "",
		meta:       meta,
	}
	return nil
}

type TocContent struct {
	Fragment        string   `json:"fragment"`
	MinutesRequired float64  `json:"minutes_required"`
	Href            string   `json:"href"`
	Depth           int      `json:"depth"`
	ID              string   `json:"id"`
	FullPath        string   `json:"full_path"`
	Order           int      `json:"order"`
	Label           string   `json:"label"`
	Filename        string   `json:"filename"`
	NaturalKey      []string `json:"natural_key"`
	MediaType       string   `json:"media_type"`
	URL             string   `json:"url"`
}

func (s *Safari) fetchTOC(id string) *map[string]TocContent {
	url := "/api/v1/book/" + id + "/flat-toc/"
	body, err := s.fetchResource(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var raw []TocContent
	err = json.Unmarshal([]byte(body), &raw)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	toc := make(map[string]TocContent)
	for _, content := range raw {
		s.books[id].toc[content.URL] = content
		toc[content.URL] = content
	}

	return &toc
}

func (s *Safari) fetchChapters(id string) error {
	// TODO: check if meta exists
	errChan := make(chan error, 1)
	resultChan := make(chan struct {
		int
		Chapter
	}, len(s.books[id].meta.Chapters))
	sem := make(chan int, 1) // four jobs at once
	var wg sync.WaitGroup
	wg.Add(len(s.books[id].meta.Chapters))
	for index, uri := range s.books[id].meta.Chapters {
		go s.fetchChapterContent(index, id, uri, sem, &wg, resultChan, errChan)
	}
	wg.Wait()
	//err := <-errChan
	close(errChan)
	close(resultChan)

	book := s.books[id]
	for r := range resultChan {
		book.chapters[r.int] = r.Chapter
	}
	s.books[id] = book
	return nil
}

func (s *Safari) fetchChapterContent(index int, id string, url string, sem chan int, wg *sync.WaitGroup, resultChan chan struct {
	int
	Chapter
}, errChan chan error) {
	defer wg.Done()
	sem <- 1
	uri := strings.Replace(url, s.baseUrl, "", -1)
	body, err := s.fetchResource(uri)
	var meta ChapterMeta
	err = json.Unmarshal([]byte(body), &meta)
	if err != nil {
		fmt.Println(err)
		return
	}
	content_url := meta.Content
	content_uri := strings.Replace(content_url, s.baseUrl, "", -1)
	content, err := s.fetchResource(content_uri)

	var chapter Chapter
	chapter.Filename = meta.Filename
	for _, v := range meta.Images {
		chapter.Images = append(chapter.Images, v.(string))
	}
	chapter.Title = meta.Title
	chapter.Content = content
	chapter.AssetBaseURL = meta.AssetBaseURL
	for _, Stylesheet := range meta.Stylesheets {
		chapter.StylesheetsURL = append(chapter.StylesheetsURL, Stylesheet.URL)
	}
	if chapToc, ok := s.books[id].toc[meta.URL]; ok {
		chapter.Order = chapToc.Order
		chapter.Id = chapToc.ID
	} else {
		chapter.Order = 0
		chapter.Id = "tocxhtmlfile"
	}

	//fmt.Println("fetch:" + id + " url:" + url)
	resultChan <- struct {
		int
		Chapter
	}{index, chapter}
	<-sem
}

func (s *Safari) fetchStylesheet(id string) error {
	book := s.books[id]
	var stylesheet string
	for _, chapter := range book.chapters {
		for _, stylesheetURL := range chapter.StylesheetsURL {
			if stylesheet == "" {
				stylesheet = stylesheetURL
			}
			if stylesheet != stylesheetURL {
				fmt.Println("an error occurred while fetching stylesheets, there are different stylesheets.")
			}
		}
	}
	book.stylesheet = stylesheet
	fmt.Println(book.stylesheet)
	s.books[id] = book
	return nil
}

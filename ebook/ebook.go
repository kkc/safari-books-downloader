package ebook

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// Ebook Chapter
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

// Ebook OebpsContent
type OebpsContent struct {
	Title   string
	Content string
	CoreCSS string
}

type JsonBook struct {
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

type ImageToFetch struct {
	BaseUrl string
	File    string
	Media   string
	Path    string
}

type Ebook struct {
	jsonBook     JsonBook
	tempBookPath string
	images       []ImageToFetch
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func prepareFolder(tempBookPath string) {
	os.MkdirAll(tempBookPath, os.ModePerm)
	os.MkdirAll(tempBookPath+"/META-INF", os.ModePerm)
	os.MkdirAll(tempBookPath+"/OEBPS", os.ModePerm)
	os.MkdirAll(tempBookPath+"/OEBPS/images", os.ModePerm)
}

func writeMimeType(tempBookPath string) {
	f, err := os.Create(tempBookPath + "/mimetype")
	check(err)
	defer f.Close()

	f.WriteString("application/epub+zip")
}

func writeContainer(tempBookPath string) {
	var container = `<?xml version="1.0" encoding="UTF-8" ?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles>
<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
</rootfiles>
</container>`

	f, err := os.Create(tempBookPath + "/META-INF/container.xml")
	check(err)
	defer f.Close()

	f.WriteString(container)
}

func NewEbook(jsonInput []byte) *Ebook {
	var jsonBook JsonBook
	err := json.Unmarshal(jsonInput, &jsonBook)
	if err != nil {
		log.Fatal(err)
	}
	tempBookPath := "books/" + jsonBook.Uuid
	prepareFolder(tempBookPath)
	writeMimeType(tempBookPath)
	writeContainer(tempBookPath)

	ebook := &Ebook{
		jsonBook:     jsonBook,
		tempBookPath: tempBookPath,
	}
	return ebook
}

// Saves the epub to the specified path
func (e *Ebook) Save(outputPath string) {
	if outputPath == "" {
		outputPath = "ebook.epub"
	}
	e.downloadImages()
	e.writeChapters()
	e.writeContentOPF()
	e.writeTOC()
	e.downloadCoverImage()
	e.writeCSS()
	e.downloadStylesheet()
	e.generateEpub(outputPath)
}

func (e *Ebook) downloadImages() {
	var images []ImageToFetch
	for _, chapter := range e.jsonBook.Chapters {
		baseUrl := chapter.AssetBaseURL
		for _, image := range chapter.Images {
			imagePath := image
			pathArray := strings.Split(image, "/")
			pathArrayLen := len(pathArray)
			if pathArrayLen > 1 {
				imagePath = pathArray[pathArrayLen-1]
			}
			//mediaType, a, err := mime.ParseMediaType(image)
			images = append(images, ImageToFetch{
				BaseUrl: baseUrl,
				File:    image,
				Media:   "image/png",
				Path:    "images/" + imagePath,
			})
		}
	}

	e.images = images

	// download images
	// TODO: wrap this as smaller function
	for _, image := range images {
		url := image.BaseUrl + image.File
		fmt.Println("fetch uri " + url)
		req, err := http.NewRequest("GET", url, nil)
		check(err)
		client := &http.Client{}
		resp, err := client.Do(req)
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err = errors.New("Error: status code != 200, actual status code '" + resp.Status + "'")
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		f, err := os.Create(e.tempBookPath + "/OEBPS/" + image.Path)
		check(err)
		defer f.Close()

		f.Write(body)
	}
}

// Write Chapters to epub file
func (e *Ebook) writeChapters() {

	var chapterTmpl = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xmlns:m="http://www.w3.org/1998/Math/MathML" xmlns:pls="http://www.w3.org/2005/01/pronunciation-lexicon" xmlns:ssml="http://www.w3.org/2001/10/synthesis" xmlns:svg="http://www.w3.org/2000/svg">
<head>
  <meta charset="UTF-8" />
  <title>{{ .Title }}</title>
  <link type="text/css" rel="stylesheet" media="all" href="style.css" />
  {{ .CoreCSS }}
</head>
<body>
  {{ .Content }}
</body>
</html>
`
	for _, chapter := range e.jsonBook.Chapters {
		var chapterContent = chapter.Content
		//TODO: replace the image source with the new local source
		for _, image := range chapter.Images {
			imagePath := image
			pathArray := strings.Split(image, "/")
			pathArrayLen := len(pathArray)
			if pathArrayLen > 1 {
				imagePath = pathArray[pathArrayLen-1]
			}
			reg := regexp.MustCompile(`[^"]*` + image)
			rep := "images/" + imagePath
			chapterContent = reg.ReplaceAllString(chapterContent, rep)
		}

		coreCSS := ""
		if e.jsonBook.Stylesheet != "" {
			coreCSS = `<link type="text/css" rel="stylesheet" media="all" href="core.css" />`
		}

		chapterContent = e.purifyHTML(chapterContent)
		c := &OebpsContent{
			Title:   chapter.Title,
			Content: chapterContent,
			CoreCSS: coreCSS,
		}

		f, err := os.Create(e.tempBookPath + "/OEBPS/" + chapter.Filename)
		check(err)
		defer f.Close()

		t := template.New("chapter")
		t, _ = t.Parse(chapterTmpl)
		err = t.Execute(f, c)
		check(err)
	}
}

// creates the content.opf file in the OEBPS directory
func (e *Ebook) writeContentOPF() {
	t := time.Now()
	date := t.UTC().Format("2006-01-02")
	isoDate := t.UTC().Format("2006-01-02T15:04:05-0700Z")

	data := struct {
		Title       string
		Uuid        string
		Language    string
		Author      string
		Cover       string
		Description string
		Publisher   string
		Stylesheet  string
		Chapters    []Chapter
		Date        string
		ISODate     string
		DateYear    int
		Creator     string
		Images      []ImageToFetch
	}{
		Title:       e.jsonBook.Title,
		Uuid:        e.jsonBook.Uuid,
		Language:    e.jsonBook.Language,
		Author:      strings.Join(e.jsonBook.Author, " "),
		Cover:       e.jsonBook.Cover,
		Description: e.jsonBook.Description,
		Publisher:   strings.Join(e.jsonBook.Publisher, ""),
		Stylesheet:  e.jsonBook.Stylesheet,
		Chapters:    e.jsonBook.Chapters,
		Date:        date,
		ISODate:     isoDate,
		DateYear:    t.UTC().Year(),
		Creator:     strings.Join(e.jsonBook.Author, " "),
		Images:      e.images,
	}

	f, err := os.Create(e.tempBookPath + "/OEBPS/content.opf")
	check(err)
	defer f.Close()

	temp := template.New("opf.tmpl")
	cwd, err := os.Getwd()
	check(err)
	temp, err = temp.ParseFiles(filepath.Join(cwd, "/ebook/opf.tmpl"))
	check(err)

	fmt.Println(data)
	err = temp.Execute(f, data)
	fmt.Println(err)
	check(err)
}

func (e *Ebook) writeTOC() {
	f, err := os.Create(e.tempBookPath + "/OEBPS/toc.ncx")
	check(err)
	defer f.Close()

	data := struct {
		Title    string
		Uuid     string
		Author   string
		Chapters []Chapter
	}{
		Title:    e.jsonBook.Title,
		Uuid:     e.jsonBook.Uuid,
		Author:   strings.Join(e.jsonBook.Author, " "),
		Chapters: e.jsonBook.Chapters,
	}
	temp := template.New("toc.ncx.tmpl")
	cwd, err := os.Getwd()
	check(err)
	temp, err = temp.ParseFiles(filepath.Join(cwd, "./ebook/toc.ncx.tmpl"))
	check(err)
	err = temp.Execute(f, data)
	check(err)

}

func (e *Ebook) downloadCoverImage() {

	out, err := os.Create(e.tempBookPath + "/OEBPS/images/cover.jpg")
	check(err)
	defer out.Close()

	resp, err := http.Get(e.jsonBook.Cover)
	check(err)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	check(err)
}

// creates the style.css file in the OEBPS directory
func (e *Ebook) writeCSS() {
	cwd, err := os.Getwd()
	check(err)
	in, err := os.Open(filepath.Join(cwd, "./ebook/style.css"))
	check(err)
	defer in.Close()

	out, err := os.Create(e.tempBookPath + "/OEBPS/style.css")
	check(err)
	defer out.Close()

	_, err = io.Copy(out, in)
	check(err)
}

func (e *Ebook) downloadStylesheet() {
	out, err := os.Create(e.tempBookPath + "/OEBPS/core.css")
	check(err)
	defer out.Close()

	fmt.Println(e.jsonBook.Stylesheet)
	resp, err := http.Get(e.jsonBook.Stylesheet)
	check(err)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	check(err)
}

// Generates and saves the epub book
func (e *Ebook) generateEpub(path string) {
	fmt.Println("Zipping temp dir to " + path)
	filesNeedToBeArchived := []string{
		e.tempBookPath + "/mimetype",
		e.tempBookPath + "/META-INF",
		e.tempBookPath + "/OEBPS",
	}
	err := zipit(filesNeedToBeArchived, path)
	check(err)
}

func (e *Ebook) purifyHTML(content string) string {
	// area,base,basefont,br,col,frame,hr,img,input,isindex,keygen,link,meta,menuitem,source,track,param,embed,wbr
	// <(\s?img[^>]*[^\/])>
	result := content
	for _, ele := range [1]string{"img"} {
		reg := regexp.MustCompile(`<(\s?` + ele + `[^>]*[^\/])>`)
		rep := "<${1} />"
		result = reg.ReplaceAllString(result, rep)
	}

	for _, ele := range [2]string{"br", "hr"} {
		reg := regexp.MustCompile(`<\s?` + ele + `[^>]*>`)
		rep := "<" + ele + "/>"
		result = reg.ReplaceAllString(result, rep)
	}
	return result
}

// rewrite it from https://gist.github.com/svett/424e6784facc0ba907ae
// TODO: can be replaced to utils.go
func zipit(sources []string, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return nil
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(source)
		}

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		})
	}

	return err
}

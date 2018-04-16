package ebook

import (
	iotil "io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEBookChapterWriter(t *testing.T) {
	content, err := iotil.ReadFile("./resp")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.writeChapters()
}

func TestPurifyHTML(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)

	input := `<img src="httpatomoreillycomsourceoreillyimages926284.jpg" alt="First Edition">`
	result := ebook.purifyHTML(input)
	assert.Equal(t, `<img src="httpatomoreillycomsourceoreillyimages926284.jpg" alt="First Edition" />`, result)

	input = `<hr>`
	result = ebook.purifyHTML(input)
	assert.Equal(t, `<hr/>`, result)
}

func TestDownloadImages(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.downloadImages()
}

func TestWriteContentOPF(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.downloadImages()
	ebook.writeContentOPF()
}

func TestWriteTOC(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.writeTOC()

}

func TestDownloadCoverImage(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.downloadCoverImage()
}

func TestWriteCSS(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.writeCSS()
}

func TestDownloadStylesheet(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.downloadStylesheet()
}

func TestGenerateEpub(t *testing.T) {
	content, err := iotil.ReadFile("./resp_sample")
	assert.NoError(t, err)

	ebook := NewEbook(content)
	ebook.downloadImages()
	ebook.writeChapters()
	ebook.writeContentOPF()
	ebook.writeTOC()
	ebook.downloadCoverImage()
	ebook.writeCSS()
	ebook.downloadStylesheet()
	ebook.generateEpub("ebook.epub")
}

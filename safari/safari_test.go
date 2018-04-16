package safari

import "testing"

//func TestSafariAuthorizeUser(t *testing.T) {
//	safari := NewSafari()
//	safari.FetchBookById("1", "", "")
//}

//func TestSafariMeta(t *testing.T) {
//	safari := NewSafari()
//	safari.authorizeUser("", "")
//	_ = safari.fetchMeta("9781449317904")
//	assert.NotNil(t, safari.books["9781449317904"].meta)
//}

//func TestSafariToc(t *testing.T) {
//	safari := NewSafari()
//	safari.authorizeUser("", "")
//	toc := safari.fetchTOC("9781449317904")
//	assert.NotNil(t, toc)
//}

//func TestSafariChapters(t *testing.T) {
//	safari := NewSafari()
//	safari.authorizeUser("", "")
//	//content, _ := safari.fetchResource("api/v1/book/9781449317904/chapter-content/ch07s02.html")
//	_ = safari.fetchMeta("9781449317904")
//	_ = safari.fetchTOC("9781449317904")
//	_ = safari.fetchChapters("9781449317904")
//	_ = safari.fetchStylesheet("9781449317904")
//	//assert.NotNil(t, safari.books["9781449317904"].meta)
//}

func TestSafariChapters(t *testing.T) {
	safari := NewSafari()
	safari.FetchBookById("9781449317904", "", "")
}

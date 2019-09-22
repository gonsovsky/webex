package extra

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"strings"
)

// Расшириние для инъектирования JScript в целевую страницу
type ExtraJS struct {
	URL string
}

func (j ExtraJS) IsTarget(request *http.Request, response *http.Response) bool{
	return j.URL != ""
}

func (j ExtraJS) Perform(request *http.Request, response *http.Response) error {
	if !strings.Contains(response.Header.Get("Content-Type"), "text/html") {
		return nil
	}
	responseText, err := ioutil.ReadAll(response.Body)
	responseBuffer := bytes.NewBuffer(responseText)
	response.Body = ioutil.NopCloser(responseBuffer)
	if err != nil {
		return err
	}
	document, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return err
	}
	payload := fmt.Sprintf("<script type='text/javascript' src='%s'></script>", j.URL)
	selection := document.
		Find("head").
		AppendHtml(payload).
		Parent()
	html, err := selection.Html()
	if err != nil {
		return err
	}
	response.Body = ioutil.NopCloser(bytes.NewBufferString(html))
	return nil
}

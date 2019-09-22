package extra

import "net/http"

// Расшириние для удаления заголовков Content-Security-Policy
// Служит для заглушки предупреждений от некоторых браузеров.
type ExtraCSP struct{}

func (j ExtraCSP) IsTarget(request *http.Request, response *http.Response) bool{
	return true
}

func (c ExtraCSP) Perform(request *http.Request, response *http.Response) error {
	response.Header.Del("Content-Security-Policy")
	return nil
}


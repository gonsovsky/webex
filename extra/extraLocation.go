package extra

import "net/http"

// Расшириние для сокрытия источника перехода
type ExtraLocation struct{}

func (j ExtraLocation) IsTarget(request *http.Request, response *http.Response) bool{
	return true
}
func (l ExtraLocation) Perform(request *http.Request, response *http.Response) error {
	location, err := response.Location()
	if err != nil {
		if err == http.ErrNoLocation {
			return nil
		}
		return err
	}
	location.Scheme = ""
	location.Host = ""
	response.Header.Set("Location", location.String())
	return nil
}


package extra

import (
	"net/http"
)

// Интерфейс расширения. Изменяет содержимое целевой страницы.
// Например: инъекция JavaScript или сбор данных форм
type Extra interface {
	// Метод определяет по адресу или др. параметрам целовой страницы
	// нужны ли дополнительные действия
	IsTarget(request *http.Request, response *http.Response) bool

	// Реагирует на или замещает содержимое целевой страницы.
	Perform(request *http.Request, response *http.Response) error
}


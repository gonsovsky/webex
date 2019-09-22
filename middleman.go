package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"webex/extra"
)

// HTTP запрос-ответ
type RoundTrip struct {
	Request  http.Request
	Response http.Response
}

// Проксирует ответы между клиентом и целевым сайтом
type MiddleMan struct {
	client *http.Client
	destination *url.URL
	extras []extra.Extra
}

// перенаправить запрос на целевой сайт
func (p *MiddleMan) passRequest(request *http.Request) (*http.Request, error) {
	dest := request.URL
	dest.Scheme = p.destination.Scheme
	dest.Host = p.destination.Host

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	req, err := http.NewRequest(request.Method, dest.String(), request.Body)
	if err != nil {
		return nil, err
	}
	for key := range request.Header {
		req.Header.Set(key, request.Header.Get(key))
	}

	if request.Referer() != "" {
		req.Header.Set("Referer", strings.Replace(request.Referer(), request.Host, p.destination.Host, -1))
	}

	if request.PostForm != nil && len(request.PostForm) >= 1 {
		req.PostForm = url.Values{}
		for p,v := range request.PostForm{
			req.PostForm[p] =v
		}
	}

	req.Header.Del("Accept-Encoding")

	return req, nil
}

// обработчик HTTP(s) запросов
func (p *MiddleMan) DoClient(
	conn net.Conn,
	roundtrips chan<- RoundTrip,
) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("Ошибка в разборе запроса от клиента:", err.Error())
		return
	}
	req, err := p.passRequest(request)
	if err != nil {
		log.Println("Ошибка формирования запроса к целевому сайту")
		return
	}

	if len(request.PostForm) > 1 {
		r := ioutil.NopCloser(bytes.NewReader([]byte(request.PostForm.Encode())))
		req.Body = r
	}
	resp, err := p.client.Do(req)
	if err != nil {
		log.Println("Ошибка обращения к целевому сайту", err.Error())
		return
	}

	//обработать ответ расширениями
	for _, extra := range p.extras {
		if extra.IsTarget(req, resp) {
			err := extra.Perform(req, resp)
			if err != nil {
				log.Println("Ошибка в пакете расширения: ", err.Error())
			}
		}
	}

	clientResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Println("Ошибка формирования ответа клиенту: ", err.Error())
		return
	}

	_, err = conn.Write(clientResponse)
	if err != nil {
		log.Println("Ошибка отправки ответа клиенту: ", err.Error())
		return
	}

	roundtrips <- RoundTrip{
		Request:  *req,
		Response: *resp,
	}
}

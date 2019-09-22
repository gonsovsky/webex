package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
	"webex/extra"
)

//Переменные конфигурации
var (
	destination    = flag.String("destination", "https://www.youtube.com/", "Целевой сайт")
	localAddr      = flag.String("localaddr", "localhost:8080", "Локальный интерфейс для запуска HTTP(S) сервера")
	proxyAddr	   = flag.String("proxy", "", "Опционально. HTTP Proxy")
	injectJS	   = flag.String("inject-js", "", "JSript для инъектирования в страницы")
	insecure       = flag.Bool  ("insecure", true, "HTTP или HTTPS")
	cert	       = flag.String("cert", "", "Если HTTPS путь к сертификату x509")
	privateKey 	   = flag.String("private-key", "", "Если HTTPS путь к приватоному ключу")
)

const (
	TimeOut = 20 * time.Second
)

func newHTTPS(address, certPath, privateKeyPath string) (net.Listener, error) {
	cer, err := tls.LoadX509KeyPair(certPath, privateKeyPath)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	return tls.Listen("tcp", address, config)
}

func newHTTP(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}

func main() {
	//http клиент для запросов к целевому сайту
	client := &http.Client{
		Timeout: TimeOut,

		// Предотвратить автоматический Редирект HTTP запросов.
		// Эту задачу выполнит сам браузер.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	//использовать прокси (или ТОР) или запросов к целевому сайту
	if *proxyAddr != "" {
		dialer, err := proxy.SOCKS5("tcp", *proxyAddr, nil, proxy.Direct)
		if err != nil {
			log.Println(err.Error())
			os.Exit(-1)
		}
		httpTransport := &http.Transport{}
		httpTransport.Dial = dialer.Dial
		client.Transport = httpTransport
	}

	//расширения для дополнительной обработки ответов целевого сайта
	extras := []extra.Extra{
		extra.ExtraCSP{},
		extra.ExtraJS{URL: *injectJS},
		extra.ExtraLocation{},
	}

	//Посредник между клиентом и целевым сайтом
	destURL, err := url.Parse(*destination)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}
	middleMan := &MiddleMan{
		client: client,
		destination: destURL,
		extras: extras,
	}

	//HTTP или HTTPS клиент
	var listenAddr string
	var server net.Listener
	if *insecure {
		server, err = newHTTP(*localAddr)
		listenAddr = fmt.Sprintf("http://%s (%s)", *localAddr, *destination)
	} else {
		server, err = newHTTPS(*localAddr, *cert, *privateKey)
		listenAddr = fmt.Sprintf("https://%s (%s)", *localAddr, *destination)
	}
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}
	log.Println("WebServer:", listenAddr)


	// Лог запрсов-ответов
	trips := make(chan RoundTrip)
	go logTrips(trips)

	//обработчик входящих запросов
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("Ошибка подключения клиента:", err.Error())
			continue
		}
		go middleMan.DoClient(conn, trips)
	}
}

func logTrips(trips <-chan RoundTrip) {
	for  t  := range trips {
		log.Println(t.Request.URL.RequestURI())
	}
}

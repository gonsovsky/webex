д./з. "Вебология"
------

Зависимости
------
 - go get golang.org/x/net/proxy
 - go get github.com/PuerkitoBio/goquery

Запуск/Отладка
------

go run main.go middleman.go, браузер http://localhost:8080

Схема работы
--------
Сервис служит посредником между целевым сайтом (--destination аргумент, defult: youtube.com)
и браузером клиента. Таким образом можно:
    - обойти Cross Frame ограничения
    - трансофрмировать содержимое страниц
    - собирать необходимые данные
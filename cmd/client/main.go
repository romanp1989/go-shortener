package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// URL сервера
const endpoint string = "http://localhost:8080/"

func main() {

	data := url.Values{}
	fmt.Println("Введите длинный URL")

	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		msg := err.Error()
		log.Fatalf(msg)
	}

	long = strings.TrimSuffix(long, "\n")
	data.Set("url", long)

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		msg := err.Error()
		log.Fatalf(msg)
	}

	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Инициализация http клиента
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		msg := err.Error()
		log.Fatalf(msg)
	}

	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)

	//Всегда закрывает тело ответа
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

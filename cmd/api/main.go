// ===== ОБЪЯВЛЕНИЕ ПАКЕТА =====
// Каждый .go файл ОБЯЗАН начинаться с объявления пакета
// package main — специальный пакет, означает "это исполняемая программа"
// Если бы написали package mylib — это была бы библиотека, а не программа
package main

// ===== АННОТАЦИИ SWAGGER =====
// Эти комментарии используются swag для генерации OpenAPI спецификации
// Формат: // @ключ значение

// @title           Payment System API
// Название API, будет отображаться в Swagger UI
// Например: "Payment System API"

// @version         1.0
// Версия API (можно использовать семантическое версионирование)

// @description     API для управления платежами
// Описание API, поддерживает многострочный текст
// @description     Поддерживает создание платежей, проверку статусов
// @description     и интеграцию с платежными шлюзами

// @termsOfService  http://swagger.io/terms/
// URL условий использования API (необязательно)

// @contact.name   API Support
// @contact.url    http://github.com/namestnikoff/payment-system
// @contact.email  support@example.com
// Контактная информация для поддержки API

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// Лицензия проекта

// @host      localhost:8080
// Адрес сервера (без http://)
// В продакшене: api.payment-system.com

// @BasePath  /
// Базовый путь для всех API endpoints
// Например: /api/v1 или просто /

// ===== ИМПОРТЫ =====
// import группирует все подключаемые библиотеки
// Скобки () позволяют импортировать несколько пакетов сразу
import (
	// "fmt" — стандартный пакет Go для форматированного ввода/вывода
	// f = format, mt = multi-type (работает с разными типами данных)
	// Используется для печати в консоль, форматирования строк
	// Аналог: print() в Python, System.out.println() в Java
	"fmt"

	// "log" — стандартный пакет для логирования
	// Добавляет timestamp (время) к каждому сообщению
	// log.Fatal() — логирует ошибку и завершает программу с кодом 1
	// В продакшене используют более продвинутые логгеры (zap, logrus)
	"log"

	// "net/http" — стандартный пакет Go для работы с HTTP
	// Позволяет создавать HTTP серверы и клиенты БЕЗ внешних библиотек
	// В Go HTTP сервер входит в стандартную библиотеку (в отличие от Python/Java)
	"net/http"

	// "encoding/json" — стандартный пакет для работы с JSON
	// encoding = кодирование/декодирование
	// Marshal = Go struct → JSON (сериализация)
	// Unmarshal = JSON → Go struct (десериализация)
	"encoding/json"

	_ "github.com/namestnikoff/payment-system/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// ===== СТРУКТУРЫ ДАННЫХ =====

// Payment представляет платеж в системе
//
// ЧТО ТАКОЕ СТРУКТУРА (struct):
// - Это способ группировки связанных данных
// - Похоже на класс без методов (методы добавляются отдельно)
// - Похоже на словарь с фиксированными ключами и типами
//
// СИНТАКСИС:
// type ИмяСтруктуры struct { поля }
type Payment struct {
	// ПОЛЯ СТРУКТУРЫ:
	// ИмяПоля ТипДанных `тег`

	// ID — уникальный идентификатор платежа
	// string = текстовая строка (как str в Python)
	// `json:"id"` = JSON тег, указывает:
	//   - При конвертации в JSON это поле будет называться "id" (маленькими)
	//   - При парсинге JSON ключ "id" попадет в это поле
	// ВАЖНО: ID с большой буквы = публичное поле (видно из других пакетов)
	//         id с маленькой = приватное (только внутри этого пакета)
	ID string `json:"id"`

	// Amount — сумма платежа
	// float64 = число с плавающей точкой, 64 бита точности
	// Для денег в продакшене лучше использовать int64 (копейки/центы)
	// Причина: float64 имеет ошибки округления (0.1 + 0.2 ≠ 0.3)
	// Пример: вместо 100.50 RUB хранить 10050 копеек
	Amount float64 `json:"amount"`

	// Currency — код валюты
	// Формат: ISO 4217 (USD, EUR, RUB, GBP и т.д.)
	// 3 буквы, всегда в верхнем регистре
	Currency string `json:"currency"`

	// Status — статус платежа
	// Возможные значения: "pending", "succeeded", "failed"
	// В продакшене лучше использовать enum (константы)
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

// ===== HTTP ОБРАБОТЧИКИ (HANDLERS) =====

// handleCreatePayment обрабатывает POST запрос для создания платежа
//
// СИГНАТУРА ФУНКЦИИ:
// func = ключевое слово объявления функции
// handleCreatePayment = имя функции (с маленькой буквы = приватная)
// (w http.ResponseWriter, r *http.Request) = параметры:
//   - w = куда писать ответ (response)
//   - r = откуда читать запрос (request), * = указатель (объясню ниже)
//
// УКАЗАТЕЛЬ (*):
// В Go данные передаются ПО ЗНАЧЕНИЮ (копируются)
// Звездочка * означает "передать ссылку, а не копию"
// Зачем: http.Request большой объект, копировать его дорого
// CreatePayment создает новый платеж
// @Summary      Создать платеж
// @Description  Создает новый платеж в системе с валидацией суммы и валюты
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payment  body      Payment  true  "Данные платежа"
// @Success      201  {object}  Payment
// @Failure      400  {string}  string  "Invalid JSON или неверные данные"
// @Failure      405  {string}  string  "Method not allowed"
// @Router       /payments [post]
func handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	// Проверяем HTTP метод
	// r.Method = строка с методом запроса ("GET", "POST", "PUT" и т.д.)
	// != означает "не равно"
	if r.Method != http.MethodPost {
		// http.Error отправляет HTTP ответ с ошибкой
		// Параметры:
		// 1. w = куда писать
		// 2. "Invalid method" = текст ошибки (будет в body ответа)
		// 3. http.StatusMethodNotAllowed = HTTP код 405
		//    (правильный код для "метод не поддерживается")
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)

		// return = прекратить выполнение функции
		// Без return код ниже выполнился бы (это ошибка!)
		return
	}

	// Создаем переменную для хранения распарсенных данных
	// var = полная форма объявления переменной
	// payment = имя переменной
	// Payment = тип (наша структура выше)
	// Значение по умолчанию: пустая структура {ID:"", Amount:0, Currency:"", Status:""}
	var payment Payment

	// Декодируем JSON из тела запроса в структуру
	//
	// json.NewDecoder(r.Body) = создает декодер, читающий из тела запроса
	// r.Body = io.Reader, поток данных (как файл)
	//
	// .Decode(&payment) = декодировать JSON → структуру
	// &payment = АДРЕС переменной payment (не копия, а именно она!)
	// Зачем &: Decode должен ИЗМЕНИТЬ payment, поэтому нужна ссылка
	//
	// ВАЖНО: Decode возвращает error
	// В Go НЕТ исключений (exceptions), вместо них — ошибки (error)
	// Если JSON невалидный, err будет содержать описание проблемы
	err := json.NewDecoder(r.Body).Decode(&payment)

	// := это "короткая форма объявления"
	// Эквивалентно: var err error = json.NewDecoder(...).Decode(...)
	// Go сам определяет тип (здесь error)

	// Проверяем, была ли ошибка при декодировании
	// nil = "ничего", "null", "нет значения"
	// err != nil означает "ошибка произошла"
	if err != nil {
		// Логируем ошибку в консоль (для разработчика/администратора)
		// log.Printf = форматированный вывод в лог с timestamp
		// "Error decoding JSON: %v" = строка формата
		// %v = подставить значение в любом формате (универсальный placeholder)
		// err = что подставить
		log.Printf("Error decoding JSON: %v", err)

		// Отправляем HTTP 400 (Bad Request) клиенту
		// "Invalid JSON" = понятное сообщение для клиента API
		// НЕ отправляем детали err клиенту (это детали реализации)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// ===== ВАЛИДАЦИЯ ДАННЫХ =====
	// КРИТИЧЕСКИ ВАЖНО ДЛЯ ФИНТЕХА!

	// Проверяем сумму платежа
	// <= 0 означает "меньше или равно нулю"
	// Нельзя принимать платежи с отрицательной/нулевой суммой
	if payment.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	// Проверяем валюту
	// payment.Currency == "" проверяет пустую строку
	// В продакшене добавили бы проверку на валидный ISO код (USD, EUR и т.д.)
	if payment.Currency == "" {
		http.Error(w, "Currency is required", http.StatusBadRequest)
		return
	}

	// ===== БИЗНЕС-ЛОГИКА =====

	// Генерируем ID платежа (в продакшене используй UUID)
	// fmt.Sprintf = форматирование строки (как f-string в Python)
	// "pay_%s" = шаблон строки
	// %s = подставить строку
	// "12345" = временное значение (в реальности: uuid.New().String())
	payment.ID = fmt.Sprintf("pay_%s", "12345")

	// Устанавливаем начальный статус
	// В реальной системе здесь был бы вызов платежного шлюза
	// (Stripe, CloudPayments и т.д.)
	payment.Status = "pending"

	// Логируем успешное создание (для мониторинга)
	// %+v = подробный вывод структуры со всеми полями
	// Вывод: {ID:pay_12345 Amount:100.5 Currency:RUB Status:pending}
	//log.Printf("Payment created: %+v", payment)
	// Замени на:
	// Логируем создание платежа с детальной информацией
	// %s = строка (string)
	// %f = число с плавающей точкой (float)
	// %q = строка в кавычках (удобно для текстовых полей)
	log.Printf("Payment created: ID=%s, Amount=%.2f %s, Status=%s, Description=%q",
		payment.ID,          // ID платежа
		payment.Amount,      // Сумма (%.2f = 2 знака после запятой)
		payment.Currency,    // Валюта
		payment.Status,      // Статус
		payment.Description) // Описание (в кавычках)

	// ===== ОТПРАВКА ОТВЕТА =====

	// Устанавливаем заголовок Content-Type
	// w.Header() = map с HTTP заголовками (как dict в Python)
	// .Set("ключ", "значение") = установить заголовок
	// "application/json" = сообщаем клиенту что отправляем JSON
	w.Header().Set("Content-Type", "application/json")

	// Устанавливаем HTTP статус код 201 (Created)
	// 201 = "ресурс успешно создан" (правильный код для POST)
	// НЕ 200, потому что 200 = "ok, но ничего не создано"
	w.WriteHeader(http.StatusCreated)

	// Кодируем структуру payment в JSON и отправляем клиенту
	// json.NewEncoder(w) = создает энкодер, пишущий в w (ResponseWriter)
	// .Encode(payment) = структуру → JSON → отправить
	// Если ошибка кодирования — игнорируем (поздно что-то менять)
	json.NewEncoder(w).Encode(payment)

	// Что увидит клиент:
	// HTTP/1.1 201 Created
	// Content-Type: application/json
	//
	// {"id":"pay_12345","amount":100.5,"currency":"RUB","status":"pending"}
}

// handleGetPayment обрабатывает GET запрос для получения статуса платежа
// В реальности здесь был бы ID в URL (например: GET /payments/pay_12345)
// Пока возвращаем заглушку (mock data)
// GetPaymentStatus получает статус платежа
// @Summary      Получить статус платежа
// @Description  Возвращает информацию о статусе платежа
// @Tags         payments
// @Produce      json
// @Success      200  {object}  Payment
// @Failure      405  {string}  string  "Method not allowed"
// @Router       /payments/status [get]
func handleGetPayment(w http.ResponseWriter, r *http.Request) {
	// Проверяем что это GET запрос
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// Создаем тестовый платеж (в реальности: запрос к БД)
	// Короткая форма инициализации структуры
	// Порядок полей не важен, можно указать только некоторые
	payment := Payment{
		ID:       "pay_12345",
		Amount:   1000.50,
		Currency: "RUB",
		Status:   "succeeded", // Платеж успешно обработан
	}

	// Отправляем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	// Для GET используем статус 200 (OK) — это стандарт
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payment)
}

// ===== ТОЧКА ВХОДА В ПРОГРАММУ =====

// main — специальная функция, выполняется при запуске программы
// Это как if __name__ == "__main__" в Python
// Должна быть в пакете main, иначе Go не запустится
func main() {
	// Выводим сообщение о запуске сервера
	// Это не обязательно, но полезно для отладки
	fmt.Println("Payment System API starting...")

	// ===== РЕГИСТРАЦИЯ МАРШРУТОВ (ROUTING) =====

	// http.HandleFunc регистрирует обработчик для URL пути
	// Параметры:
	// 1. "/payments" = URL путь (pattern)
	//    Запросы на http://localhost:8080/payments попадут сюда
	// 2. handleCreatePayment = функция-обработчик
	//    Передаем ФУНКЦИЮ (не вызываем её!)
	//    Без скобок () — это важно!
	http.HandleFunc("/payments", handleCreatePayment)

	// Еще один маршрут для получения платежа
	// В реальности путь был бы "/payments/{id}"
	// Для простоты пока статичный путь
	http.HandleFunc("/payments/status", handleGetPayment)

	// ===== ЗАПУСК HTTP СЕРВЕРА =====

	// http.ListenAndServe запускает HTTP сервер
	// Параметры:
	// 1. ":8080" = адрес и порт
	//    : без IP = слушать на всех сетевых интерфейсах (0.0.0.0)
	//    8080 = номер порта (можно любой от 1024 до 65535)
	// 2. nil = использовать DefaultServeMux (роутер по умолчанию)
	//    Все маршруты из HandleFunc идут туда
	//
	// ЭТА ФУНКЦИЯ БЛОКИРУЮЩАЯ:
	// После её вызова программа "зависает" и обрабатывает запросы
	// Код после ListenAndServe выполнится только при ошибке или остановке
	//
	// ВОЗВРАЩАЕТ ERROR:
	// Если сервер не смог запуститься (порт занят и т.д.)

	// Swagger UI будет доступен по адресу http://localhost:8080/swagger/
	// httpSwagger.WrapHandler = функция которая создает HTTP обработчик для Swagger UI
	// Она читает спецификацию (которую мы сгенерируем командой swag init)
	// и отображает интерактивную документацию в браузере
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	log.Println("Server is running on http://localhost:8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	err := http.ListenAndServe(":8080", nil)

	// Если мы здесь — значит сервер упал
	// log.Fatal логирует ошибку и вызывает os.Exit(1)
	// Программа завершается с кодом ошибки 1
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// ===== ЧТО ПРОИСХОДИТ ПРИ ЗАПУСКЕ =====
//
// 1. Go компилирует код в исполняемый файл (бинарник)
// 2. Запускает функцию main()
// 3. Регистрирует маршруты /payments и /payments/status
// 4. Запускает HTTP сервер на порту 8080
// 5. Ждет входящих HTTP запросов
// 6. При запросе на /payments вызывает handleCreatePayment
// 7. При запросе на /payments/status вызывает handleGetPayment
//
// ===== ПРИМЕР ИСПОЛЬЗОВАНИЯ =====
//
// Создание платежа:
// curl -X POST http://localhost:8080/payments \
//   -H "Content-Type: application/json" \
//   -d '{"amount": 1000.50, "currency": "RUB"}'
//
// Ответ:
// {"id":"pay_12345","amount":1000.5,"currency":"RUB","status":"pending"}
//
// Получение статуса:
// curl http://localhost:8080/payments/status
//
// Ответ:
// {"id":"pay_12345","amount":1000.5,"currency":"RUB","status":"succeeded"}

package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"qwalletrestapi/internal/qwalletstore"
	"qwalletrestapi/internal/storage/postges"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type HTTPHandlers struct {
	database *postges.PostgresStore
}

func NewHTTPHandlers(database *postges.PostgresStore) *HTTPHandlers {
	return &HTTPHandlers{
		database: database,
	}
}

/*
{
    "from": "uuuuuuu",
    "to":"fgfdgdfg",
    "amount":100
}
*/

/*
pattern: /api/send
method:  POST
info:    JSON in HTTP request body

succeed:
  - status code:   201
  - response body: JSON

failed:
  - status code:   400, 409, 500, ...
  - response body: JSON with error + time
*/

func (h *HTTPHandlers) HandleSend(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Читаем и валидируем JSON
	var dto TransactionDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		errDTO := ErrorDTO{
			Message: "Invalid JSON format: " + err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}

	// 2. Валидация данных
	if dto.FromAddress == "" || dto.ToAddress == "" {
		errDTO := ErrorDTO{
			Message: "From and to addresses are required",
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}

	if dto.Amount <= 0 {
		errDTO := ErrorDTO{
			Message: "Amount must be positive",
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}

	// 3. Проверяем существование кошельков
	fromWallet, err := h.database.GetWallet(dto.FromAddress)
	if err != nil {
		errDTO := ErrorDTO{
			Message: "Sender wallet not found: " + err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusNotFound)
		return
	}

	_, err = h.database.GetWallet(dto.ToAddress)
	if err != nil {
		errDTO := ErrorDTO{
			Message: "Receiver wallet not found: " + err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusNotFound)
		return
	}

	// 4. Проверяем достаточность баланса
	if fromWallet.Balance < dto.Amount {
		errDTO := ErrorDTO{
			Message: fmt.Sprintf("Insufficient funds: current balance %.2f, required %.2f",
				fromWallet.Balance, dto.Amount),
			Time: time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}

	// 5. Выполняем перевод средств
	err = h.database.SendWallet(dto.FromAddress, dto.ToAddress, dto.Amount)
	if err != nil {
		errDTO := ErrorDTO{
			Message: "Transfer failed: " + err.Error(),
			Time:    time.Now(),
		}
		http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
		return
	}

	// 6. Создаем запись о транзакции
	transaction := &qwalletstore.Transaction{
		FromAddress: dto.FromAddress,
		ToAddress:   dto.ToAddress,
		Amount:      dto.Amount,
	}

	err = h.database.CreateTransaction(transaction)
	if err != nil {
		// Логируем ошибку, но не прерываем ответ, т.к. деньги уже переведены
		log.Printf("Warning: failed to create transaction record: %v", err)
	}

	// 7. Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":        "Transfer completed successfully",
		"transaction_id": transaction.ID,
		"created_at":     transaction.CreatedAt,
		"from":           transaction.FromAddress,
		"to":             transaction.ToAddress,
		"amount":         transaction.Amount,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

/*
pattern: /api/transactions?count=N
method:  GET
info:    pattern

succeed:
  - status code: 200 Ok
  - response body: JSON represented found task

failed:
  - status code: 400, 404, 500, ...
  - response body: JSON with error + time
*/

func (h *HTTPHandlers) HandleGetLast(w http.ResponseWriter, r *http.Request) {

	count, err := strconv.Atoi(r.URL.Query()["count"][0])
	fmt.Println(count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		msg := "fail to read HTTP body:" + err.Error()
		fmt.Println(msg)
		_, err = w.Write([]byte(msg))
		if err != nil {
			fmt.Println("fail to write HTTP response:", err)
		}
		return
	}

	h.database.GetLast(count)

}

/*
pattern: /api/wallet/{address}/balance,
method:  GET
info:    pattern

succeed:
  - status code: 200 Ok
  - response body: JSON represented found task

failed:
  - status code: 400, 404, 500, ...
  - response body: JSON with error + time
*/

func (h *HTTPHandlers) HandleGetBalance(w http.ResponseWriter, r *http.Request) {

	_, err := io.ReadAll(r.Body) //

	vars := mux.Vars(r)
	address := vars["address"]

	//qwallet := qwalletstore.Qwallet{}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		msg := "fail to read HTTP body:" + err.Error()
		fmt.Println(msg)
		_, err = w.Write([]byte(msg))
		if err != nil {
			fmt.Println("fail to write HTTP response:", err)
		}
		return
	}

	//qwallet, err := postges.GetBalance(address)

	qwallet, err := h.database.GetWallet(address)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(qwallet)

}

//http://localhost:8080/api/wallet/d6a6a923049b728b1534e37a0c1c3f21/balance

package postges

import (
	"database/sql"
	"fmt"
	"log"
	"qwalletrestapi/internal/qwalletstore"

	_ "github.com/lib/pq"
)

// PostgresStore хранит подключение к базе данных
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore создаёт новый экземпляр хранилища с подключением к БД
func NewPostgresStore(storagePath string) (*PostgresStore, error) {
	// Подключение к базе данных
	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Инициализация кошельков (если таблица пуста)
	if err := seedWallet(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to seed wallets: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

// Close закрывает соединение с базой данных
func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// seedWallet создаёт 10 случайных кошельков, если таблица пуста
func seedWallet(db *sql.DB) error {
	// Проверяем, есть ли хоть один кошелёк
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM qwallets").Scan(&count)
	if err != nil {
		return fmt.Errorf("count wallets: %w", err)
	}
	if count > 0 {
		log.Println("⚡️ Кошельки уже существуют, пропускаем")
		return nil
	}

	// Если таблица пуста — создаём 10 случайных кошельков
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO qwallets(address, balance) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := 0; i < 10; i++ {
		addr := qwalletstore.GenerateRandomAddress()
		_, err := stmt.Exec(addr, 100.0) // 100 у.е. на старте
		if err != nil {
			return fmt.Errorf("add wallet: %w", err)
		}
		log.Printf("✅ Создан кошелёк %s с балансом %.2f", addr, 100.0)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Println("Создано 10 кошельков")
	return nil
}

func (s *PostgresStore) GetWallet(address string) (*qwalletstore.Qwallet, error) {
	qwallet := qwalletstore.Qwallet{}

	err := s.db.QueryRow(
		`SELECT address, balance FROM qwallets WHERE address = $1`,
		address,
	).Scan(&qwallet.Address, &qwallet.Balance)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("кошелек с адресом '%s' не найден", address)
		}
		return nil, fmt.Errorf("ошибка базы данных: %w", err)
	}

	return &qwallet, nil
}

// CreateTransaction создаёт новую транзакцию
func (s *PostgresStore) SendWallet(from, to string, amount float64) error {

	dbTx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("Ошибка в создании транзакции: %w", err)
	}
	defer dbTx.Rollback()

	// Проверка с FOR UPDATE для блокировки строки
	var currentBalance float64
	err = dbTx.QueryRow(
		"SELECT balance FROM qwallets WHERE address = $1 FOR UPDATE",
		from,
	).Scan(&currentBalance)

	if err != nil {
		return fmt.Errorf("failed to get sender balance: %w", err)
	}

	// ДВОЙНАЯ ПРОВЕРКА - защита от race condition
	if currentBalance < amount {
		return fmt.Errorf("недостаточно средств: текущий баланс %.2f, требуется %.2f",
			currentBalance, amount)
	}

	// Обновление балансов
	_, err = dbTx.Exec(
		"UPDATE qwallets SET balance = balance - $1 WHERE address = $2",
		amount, from,
	)
	if err != nil {
		return fmt.Errorf("ошибка при пополнении перевода %w",
			err)
	}

	_, err = dbTx.Exec(
		"UPDATE qwallets SET balance = balance + $1 WHERE address = $2", amount, to,
	)

	if err != nil {
		return fmt.Errorf("ошибка при выполнении депозита %w",
			err)
	}

	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("Ошибка в завершении транзакции %w", &err)
	}

	return nil
}

// CreateTransaction создает запись о транзакции в базе данных
func (s *PostgresStore) CreateTransaction(tx *qwalletstore.Transaction) error {
	err := s.db.QueryRow(
		`INSERT INTO transactions (from_address, to_address, amount) 
         VALUES ($1, $2, $3) 
         RETURNING id, created_at`,
		tx.FromAddress, tx.ToAddress, tx.Amount,
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (s *PostgresStore) count_txt() int {
	rows, err := s.db.Query("SELECT COUNT(*) FROM transactions")
	if err != nil {
		log.Fatal(err)
	}
	var count int
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Fatal(err)
		}
	}
	return count
}

func (s *PostgresStore) GetLast(tx_cnt int) {

	if tx_cnt > s.count_txt() {
		fmt.Println("в базе нет столько переводов")
	}

	rows, err := s.db.Query(
		`SELECT TOP 5 *
		FROM transactions
		ORDER BY created_at`,
	)

	defer rows.Close()
	if err != nil {
		fmt.Errorf("failed to select transaction: %w", err)
	}

	for rows.Next() {

		tx := qwalletstore.Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromAddress, &tx.ToAddress, &tx.Amount, &tx.CreatedAt)

		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(tx)
	}

}

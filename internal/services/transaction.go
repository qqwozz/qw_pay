package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/models"
)

// exchangeRates — фиксированные курсы валют (для MVP).
// Ключ формата: "FROM_TO", например "RUB_USD".
var exchangeRates = map[string]float64{
	"RUB_USD": 0.011,
	"USD_RUB": 90.91,
	"RUB_EUR": 0.010,
	"EUR_RUB": 100.0,
	"USD_EUR": 0.92,
	"EUR_USD": 1.09,
}

// TransactionService отвечает за создание и хранение переводов.
type TransactionService struct {
	db       *pgxpool.Pool
	accounts *AccountService
}

// NewTransactionService создаёт новый экземпляр TransactionService.
func NewTransactionService(db *pgxpool.Pool, accounts *AccountService) *TransactionService {
	return &TransactionService{db: db, accounts: accounts}
}

// Create выполняет перевод между счетами с проверками:
//   - идемпотентность (повторный запрос с тем же ключом возвращает существующий результат);
//   - существование и активность счетов;
//   - лимиты (максимальная сумма, дневной лимит);
//   - конвертация при разных валютах;
//   - optimistic locking при обновлении балансов.
func (s *TransactionService) Create(ctx context.Context, fromID, toID uuid.UUID, amount float64, idempotencyKey string) (*models.Transaction, error) {
	// 1. Проверка идемпотентности — если транзакция уже существует, возвращаем её
	var existing models.Transaction
	err := s.db.QueryRow(ctx,
		`SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency,
		        source_currency, exchange_rate_used, status, created_at
		 FROM transactions WHERE idempotency_key=$1`, idempotencyKey,
	).Scan(&existing.ID, &existing.IdempotencyKey, &existing.FromAccountID, &existing.ToAccountID,
		&existing.Amount, &existing.Currency, &existing.SourceCurrency, &existing.ExchangeRate,
		&existing.Status, &existing.CreatedAt)
	if err == nil {
		return &existing, nil
	}

	// 2. Проверка счетов
	from, err := s.accounts.GetByID(ctx, fromID)
	if err != nil {
		return nil, fmt.Errorf("source account not found")
	}
	to, err := s.accounts.GetByID(ctx, toID)
	if err != nil {
		return nil, fmt.Errorf("target account not found")
	}

	// 3. Проверка статуса счетов
	if from.Status != models.StatusActive || to.Status != models.StatusActive {
		return nil, fmt.Errorf("account is blocked")
	}

	// 4. Проверка лимитов
	if amount > config.C.MaxTransferAmount {
		return nil, fmt.Errorf("amount exceeds max transfer limit")
	}
	if fromID == toID {
		return nil, fmt.Errorf("cannot transfer to the same account")
	}

	// 5. Проверка дневного лимита
	dailySum, err := s.accounts.GetDailyTransferSum(ctx, fromID)
	if err != nil {
		return nil, err
	}
	if dailySum+amount > config.C.DailyLimit {
		return nil, fmt.Errorf("daily transfer limit exceeded")
	}

	// 6. Определение курса конвертации (если валюты разные)
	var exchangeRate *float64
	sourceCurrency := string(from.Currency)
	targetCurrency := string(to.Currency)

	if sourceCurrency != targetCurrency {
		key := sourceCurrency + "_" + targetCurrency
		rate, ok := exchangeRates[key]
		if !ok {
			return nil, fmt.Errorf("no exchange rate for %s to %s", sourceCurrency, targetCurrency)
		}
		exchangeRate = &rate
	}

	// 7. Выполнение в транзакции БД (ACID)
	dbTx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer dbTx.Rollback(ctx)

	// 7.1. Вставка записи о переводе
	var transaction models.Transaction
	err = dbTx.QueryRow(ctx,
		`INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, currency, source_currency, exchange_rate_used, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'EXECUTED')
		 RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, source_currency, exchange_rate_used, status, created_at`,
		idempotencyKey, fromID, toID, amount, targetCurrency, sourceCurrency, exchangeRate,
	).Scan(&transaction.ID, &transaction.IdempotencyKey, &transaction.FromAccountID, &transaction.ToAccountID,
		&transaction.Amount, &transaction.Currency, &transaction.SourceCurrency, &transaction.ExchangeRate,
		&transaction.Status, &transaction.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 7.2. Списание со счёта отправителя (optimistic lock через version)
	tag, err := dbTx.Exec(ctx,
		`UPDATE accounts SET balance=balance-$1, version=version+1, updated_at=NOW()
		 WHERE id=$2 AND version=$3 AND status='ACTIVE'`,
		amount, fromID, from.Version,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("optimistic lock conflict on source account")
	}

	// 7.3. Зачисление на счёт получателя (optimistic lock через version)
	tag, err = dbTx.Exec(ctx,
		`UPDATE accounts SET balance=balance+$1, version=version+1, updated_at=NOW()
		 WHERE id=$2 AND version=$3 AND status='ACTIVE'`,
		amount, toID, to.Version,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("optimistic lock conflict on target account")
	}

	// 7.4. Фиксация транзакции
	if err := dbTx.Commit(ctx); err != nil {
		return nil, err
	}

	log.Printf("Transaction executed: id=%s from=%s to=%s amount=%.2f %s → %s",
		transaction.ID, fromID, toID, amount, sourceCurrency, targetCurrency)
	return &transaction, nil
}

// ListByUser возвращает список транзакций пользователя с пагинацией.
func (s *TransactionService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Transaction, int, error) {
	offset := (page - 1) * pageSize

	// Получаем транзакции, связанные со счетами пользователя
	rows, err := s.db.Query(ctx,
		`SELECT t.id, t.idempotency_key, t.from_account_id, t.to_account_id, t.amount,
		        t.currency, t.source_currency, t.exchange_rate_used, t.status, t.created_at
		 FROM transactions t
		 JOIN accounts a ON a.id = t.from_account_id OR a.id = t.to_account_id
		 WHERE a.user_id = $1
		 GROUP BY t.id
		 ORDER BY t.created_at DESC
		 OFFSET $2 LIMIT $3`,
		userID, offset, pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.IdempotencyKey, &tx.FromAccountID, &tx.ToAccountID,
			&tx.Amount, &tx.Currency, &tx.SourceCurrency, &tx.ExchangeRate, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}

	// Общее количество для пагинации
	var total int
	err = s.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT t.id)
		 FROM transactions t
		 JOIN accounts a ON a.id = t.from_account_id OR a.id = t.to_account_id
		 WHERE a.user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

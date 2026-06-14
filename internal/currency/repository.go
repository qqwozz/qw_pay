package currency

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExchangeRate struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Source       string  `json:"source"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	er := &ExchangeRate{}
	err := r.db.QueryRow(ctx,
		`SELECT from_currency, to_currency, rate, source
		 FROM exchange_rates WHERE from_currency=$1 AND to_currency=$2`, from, to,
	).Scan(&er.FromCurrency, &er.ToCurrency, &er.Rate, &er.Source)
	if err != nil {
		return nil, err
	}
	return er, nil
}

func (r *Repository) GetAllRates(ctx context.Context) ([]ExchangeRate, error) {
	rows, err := r.db.Query(ctx,
		`SELECT from_currency, to_currency, rate, source
		 FROM exchange_rates ORDER BY from_currency, to_currency`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []ExchangeRate
	for rows.Next() {
		var er ExchangeRate
		if err := rows.Scan(&er.FromCurrency, &er.ToCurrency, &er.Rate, &er.Source); err != nil {
			return nil, err
		}
		rates = append(rates, er)
	}
	return rates, nil
}

func (r *Repository) UpsertRate(ctx context.Context, from, to string, rate float64, source string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO exchange_rates (from_currency, to_currency, rate, source)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (from_currency, to_currency)
		 DO UPDATE SET rate = EXCLUDED.rate, source = EXCLUDED.source`,
		from, to, rate, source)
	return err
}

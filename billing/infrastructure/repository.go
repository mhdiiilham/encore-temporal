package infrastructure

import (
	"context"
	"fmt"

	"encore.app/billing/domain"
	"encore.dev/storage/sqldb"
)

// repository implements domain.Repository interface
type repository struct {
	db *sqldb.Database
}

// NewRepository creates a new repository instance
func NewRepository(db *sqldb.Database) domain.Repository {
	return &repository{db: db}
}

// Bill operations
func (r *repository) GetBill(ctx context.Context, billingID string) (domain.Bill, error) {
	const q = `
	SELECT id, billing_id, status, currency, total, created_at, closed_at
	FROM bills
	WHERE billing_id = $1
	`

	var bill domain.Bill
	err := r.db.QueryRow(ctx, q, billingID).Scan(
		&bill.ID,
		&bill.BillingID,
		&bill.Status,
		&bill.Currency,
		&bill.Total,
		&bill.CreatedAt,
		&bill.ClosedAt,
	)
	if err != nil {
		return domain.Bill{}, fmt.Errorf("failed to get bill: %w", err)
	}

	items, err := r.GetItemsByBillID(ctx, billingID)
	if err != nil {
		return domain.Bill{}, fmt.Errorf("failed to get items: %w", err)
	}
	bill.Items = items

	exchange, err := r.GetExchangeByBillID(ctx, billingID)
	if err == nil {
		bill.Conversion = exchange
	}

	return bill, nil
}

func (r *repository) SaveBill(ctx context.Context, bill *domain.Bill) error {
	const q = `
	INSERT INTO bills (billing_id, status, currency, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (billing_id) DO UPDATE
	SET status = EXCLUDED.status,
	    currency = EXCLUDED.currency,
	    created_at = EXCLUDED.created_at
	RETURNING id
	`

	err := r.db.QueryRow(ctx, q,
		bill.BillingID,
		bill.Status,
		bill.Currency,
		bill.CreatedAt,
	).Scan(&bill.ID)

	if err != nil {
		return err
	}
	return nil
}

func (r *repository) UpdateBill(ctx context.Context, bill *domain.Bill) error {
	const q = `
	UPDATE bills
	SET status = $2, currency = $3, total = $4, closed_at = $5
	WHERE billing_id = $1
	`

	_, err := r.db.Exec(ctx, q,
		bill.BillingID,
		bill.Status,
		bill.Currency,
		bill.Total,
		bill.ClosedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update bill: %w", err)
	}
	return nil
}

func (r *repository) DeleteBill(ctx context.Context, billingID string) error {
	const q = `DELETE FROM bills WHERE billing_id = $1`

	_, err := r.db.Exec(ctx, q, billingID)
	if err != nil {
		return fmt.Errorf("failed to delete bill: %w", err)
	}
	return nil
}

func (r *repository) SaveItem(ctx context.Context, item *domain.Item) error {
	const q = `
	INSERT INTO bill_items (bill_id, name, price, idemp_key)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (idemp_key)
	DO UPDATE SET
		name = EXCLUDED.name,
		price = EXCLUDED.price
	RETURNING id
	`

	err := r.db.QueryRow(ctx, q,
		item.BillingID,
		item.Name,
		item.Price,
		item.IdempotencyKey,
	).Scan(&item.ID)

	if err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}
	return nil
}

func (r *repository) GetItemsByBillID(ctx context.Context, billID string) ([]domain.Item, error) {
	const q = `
	SELECT id, bill_id, name, price, idemp_key
	FROM bill_items
	WHERE bill_id = $1
	ORDER BY id
	`

	rows, err := r.db.Query(ctx, q, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.BillingID, &item.Name, &item.Price, &item.IdempotencyKey); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

func (r *repository) UpdateItem(ctx context.Context, item *domain.Item) error {
	const q = `
	UPDATE bill_items
	SET name = $2, price = $3
	WHERE id = $1
	`

	_, err := r.db.Exec(ctx, q, item.ID, item.Name, item.Price)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

func (r *repository) DeleteItem(ctx context.Context, itemID int64) error {
	const q = `DELETE FROM bill_items WHERE id = $1`

	_, err := r.db.Exec(ctx, q, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

func (r *repository) SaveExchange(ctx context.Context, bill *domain.Bill) error {
	const q = `
	INSERT INTO bill_exchanges (bill_id, base_currency, target_currency, rate, total)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`

	err := r.db.QueryRow(ctx, q,
		bill.Conversion.BillID,
		bill.Conversion.BaseCurrency,
		bill.Conversion.TargetCurrency,
		bill.Conversion.Rate,
		bill.Conversion.Total,
	).Scan(&bill.Conversion.ID)

	if err != nil {
		return fmt.Errorf("failed to save exchange: %w", err)
	}
	return nil
}

func (r *repository) GetExchangeByBillID(ctx context.Context, billID string) (domain.BillExchange, error) {
	const q = `
	SELECT id, bill_id, base_currency, target_currency, rate, total
	FROM bill_exchanges
	WHERE bill_id = $1
	`

	var exchange domain.BillExchange
	err := r.db.QueryRow(ctx, q, billID).Scan(
		&exchange.ID,
		&exchange.BillID,
		&exchange.BaseCurrency,
		&exchange.TargetCurrency,
		&exchange.Rate,
		&exchange.Total,
	)

	if err != nil {
		return domain.BillExchange{}, fmt.Errorf("failed to get exchange: %w", err)
	}
	return exchange, nil
}

func (r *repository) UpdateExchange(ctx context.Context, exchange *domain.BillExchange) error {
	const q = `
	UPDATE bill_exchanges
	SET base_currency = $2, target_currency = $3, rate = $4, total = $5
	WHERE id = $1
	`

	_, err := r.db.Exec(ctx, q,
		exchange.ID,
		exchange.BaseCurrency,
		exchange.TargetCurrency,
		exchange.Rate,
		exchange.Total,
	)

	if err != nil {
		return fmt.Errorf("failed to update exchange: %w", err)
	}
	return nil
}

func (r *repository) DeleteExchange(ctx context.Context, exchangeID int64) error {
	const q = `DELETE FROM bill_exchanges WHERE id = $1`

	_, err := r.db.Exec(ctx, q, exchangeID)
	if err != nil {
		return fmt.Errorf("failed to delete exchange: %w", err)
	}
	return nil
}

func (r *repository) CloseBilling(ctx context.Context, billing domain.Bill) error {
	const q = `
	UPDATE bills
	SET status = 'CLOSED',
	    closed_at = now(),
	    total = (
	      SELECT COALESCE(SUM(price), 0)
	      FROM bill_items
	      WHERE bill_items.bill_id = bills.billing_id
	    )
	WHERE billing_id = $1
	  AND status = 'OPEN'
	RETURNING id, billing_id, status, currency, total, created_at, closed_at
	`

	err := r.db.QueryRow(ctx, q, billing.BillingID).Scan(
		&billing.ID,
		&billing.BillingID,
		&billing.Status,
		&billing.Currency,
		&billing.Total,
		&billing.CreatedAt,
		&billing.ClosedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

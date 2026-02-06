package repositories

import (
	"database/sql"
	"kasir-api/models"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetDailySalesReport retrieves sales report for today
func (r *ReportRepository) GetDailySalesReport() (*models.SalesReport, error) {
	// Get current date from database (which is already in Asia/Jakarta timezone)
	var today string
	err := r.db.QueryRow("SELECT CURRENT_DATE::text").Scan(&today)
	if err != nil {
		return nil, err
	}

	startOfDay := today + " 00:00:00"
	endOfDay := today + " 23:59:59"

	return r.GetSalesReportByDateRange(startOfDay, endOfDay)
}

// GetSalesReportByDateRange retrieves sales report for a specific date range
func (r *ReportRepository) GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error) {
	report := &models.SalesReport{}

	// Get total revenue and transaction count
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(*) as total_transaksi
		FROM transactions
		WHERE created_at >= $1 AND created_at <= $2
			AND deleted_at IS NULL
	`

	err := r.db.QueryRow(query, startDate, endDate).Scan(&report.TotalRevenue, &report.TotalTransaksi)
	if err != nil {
		return nil, err
	}

	// Get top selling product
	topProductQuery := `
		SELECT 
			p.name,
			SUM(td.quantity) as qty_terjual
		FROM transaction_details td
		INNER JOIN transactions t ON td.transaction_id = t.id
		INNER JOIN product p ON td.product_id = p.id
		WHERE t.created_at >= $1 AND t.created_at <= $2
			AND t.deleted_at IS NULL
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`

	var topProduct models.TopProduct
	err = r.db.QueryRow(topProductQuery, startDate, endDate).Scan(&topProduct.Nama, &topProduct.QtyTerjual)
	if err == sql.ErrNoRows {
		// No transactions in this period, return report with null top product
		return report, nil
	}
	if err != nil {
		return nil, err
	}

	report.ProdukTerlaris = &topProduct
	return report, nil
}

package postgresinventory

import (
	"context"
	"database/sql"

	inventoryapp "github.com/example/crud/internal/application/inventory"
	domaininventory "github.com/example/crud/internal/domain/inventory"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) inventoryapp.Repository {
	return &repository{db: db}
}

func (r *repository) ListByBranch(ctx context.Context, companyID, branchID int64) ([]domaininventory.Item, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT bi.company_id, bi.branch_id, bi.product_id, p.sku, p.name, p.category, p.brand,
			bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2
		ORDER BY p.name, p.id`,
		companyID,
		branchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domaininventory.Item, 0)
	for rows.Next() {
		var item domaininventory.Item
		if err := rows.Scan(
			&item.CompanyID,
			&item.BranchID,
			&item.ProductID,
			&item.ProductSKU,
			&item.ProductName,
			&item.Category,
			&item.Brand,
			&item.StockOnHand,
			&item.ReservedStock,
			&item.AvailableStock,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) SuggestSources(ctx context.Context, companyID, destinationBranchID, productID int64, quantity int) ([]domaininventory.SourceCandidate, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`WITH destination_branch AS (
			SELECT latitude, longitude
			FROM branches
			WHERE company_id = $1 AND id = $2
		)
		SELECT
			b.company_id,
			b.id,
			b.code,
			b.name,
			b.city,
			b.region,
			b.latitude,
			b.longitude,
			bi.product_id,
			(bi.stock_on_hand - bi.reserved_stock) AS available_stock,
			CASE
				WHEN db.latitude IS NULL OR db.longitude IS NULL OR b.latitude IS NULL OR b.longitude IS NULL THEN NULL
				ELSE 6371 * acos(
					LEAST(1, GREATEST(-1,
						cos(radians(db.latitude)) * cos(radians(b.latitude)) * cos(radians(b.longitude) - radians(db.longitude)) +
						sin(radians(db.latitude)) * sin(radians(b.latitude))
					))
				)
			END AS distance_km
		FROM branch_inventory bi
		INNER JOIN branches b ON b.id = bi.branch_id AND b.company_id = bi.company_id
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		LEFT JOIN destination_branch db ON TRUE
		WHERE bi.company_id = $1
			AND bi.product_id = $3
			AND bi.branch_id <> $2
			AND b.is_active = TRUE
			AND (bi.stock_on_hand - bi.reserved_stock) >= $4
		ORDER BY
			CASE WHEN db.latitude IS NULL OR db.longitude IS NULL OR b.latitude IS NULL OR b.longitude IS NULL THEN 1 ELSE 0 END,
			distance_km ASC NULLS LAST,
			available_stock DESC,
			b.id ASC`,
		companyID,
		destinationBranchID,
		productID,
		quantity,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]domaininventory.SourceCandidate, 0)
	for rows.Next() {
		var (
			candidate  domaininventory.SourceCandidate
			latitude   sql.NullFloat64
			longitude  sql.NullFloat64
			distanceKm sql.NullFloat64
		)
		if err := rows.Scan(
			&candidate.CompanyID,
			&candidate.BranchID,
			&candidate.BranchCode,
			&candidate.BranchName,
			&candidate.City,
			&candidate.Region,
			&latitude,
			&longitude,
			&candidate.ProductID,
			&candidate.AvailableStock,
			&distanceKm,
		); err != nil {
			return nil, err
		}
		if latitude.Valid {
			candidate.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			candidate.Longitude = &longitude.Float64
		}
		if distanceKm.Valid {
			candidate.DistanceKilometers = &distanceKm.Float64
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

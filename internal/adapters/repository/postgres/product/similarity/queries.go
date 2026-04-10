package productsimilaritypg

import (
	"context"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
)

func (s *Store) FindNeighbors(ctx context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]productdto.NeighborOutput, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		`SELECT
			p.id,
			p.sku,
			p.name,
			p.description,
			p.category,
			p.brand,
			p.price_cents,
			p.currency,
			(source.embedding <=> p.embedding) AS distance
		FROM products AS source
		JOIN products AS p
			ON p.company_id = source.company_id
			AND p.id <> source.id
			AND p.embedding IS NOT NULL
		WHERE source.id = $1
			AND source.company_id = $2
			AND source.embedding IS NOT NULL
			AND (1 - (source.embedding <=> p.embedding)) >= $3
		ORDER BY source.embedding <=> p.embedding ASC
		LIMIT $4`,
		sourceProductID,
		companyID,
		minSimilarity,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	neighbors := make([]productdto.NeighborOutput, 0)
	for rows.Next() {
		var neighbor productdto.NeighborOutput
		if err := rows.Scan(
			&neighbor.ProductID,
			&neighbor.SKU,
			&neighbor.Name,
			&neighbor.Description,
			&neighbor.Category,
			&neighbor.Brand,
			&neighbor.PriceCents,
			&neighbor.Currency,
			&neighbor.Distance,
		); err != nil {
			return nil, err
		}
		neighbors = append(neighbors, neighbor)
	}

	return neighbors, rows.Err()
}

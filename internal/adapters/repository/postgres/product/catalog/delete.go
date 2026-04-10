package productcatalogpg

import (
	"context"

	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
)

func (s *Store) Delete(ctx context.Context, id int64) error {
	result, err := s.DB.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return productapp.ErrNotFound
	}

	return nil
}

package salecommand

import (
	saledto "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/dto"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

func toDomainItems(inputs []saledto.CreateItemInput) []domainsale.Item {
	items := make([]domainsale.Item, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, domainsale.Item{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		})
	}

	return items
}

func toOutput(sale domainsale.Sale) saledto.Output {
	items := make([]saledto.ItemOutput, 0, len(sale.Items))
	for _, item := range sale.Items {
		items = append(items, saledto.ItemOutput{
			ProductID:      item.ProductID,
			ProductSKU:     item.ProductSKU,
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			SubtotalCents:  item.SubtotalCents,
		})
	}

	return saledto.Output{
		ID:               sale.ID,
		CompanyID:        sale.CompanyID,
		BranchID:         sale.BranchID,
		CreatedByUserID:  sale.CreatedByUserID,
		TotalAmountCents: sale.TotalAmountCents,
		Items:            items,
		CreatedAt:        sale.CreatedAt,
	}
}

package cart

import (
	"fmt"

	"github.com/quanghia24/mySmartHome/types"
)

func getCartItemsID(items []types.CartItem) ([]int, error) {
	productIds := make([]int, len(items))
	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for the product %d", item.ProductID)
		}
		productIds[i] = item.ProductID
	}

	return productIds, nil
}

func (h *Handler) createOrder(ps []types.Product, items []types.CartItem, userID int) (int, float64, error) {
	// having to loop through product list many time,
	// it's best to create a map for fast access
	productMap := make(map[int]types.Product)
	for _, product := range ps {
		productMap[product.ID] = product
	}

	// check if all products are in stock
	if err := checkIfCartIsInStock(items, productMap); err != nil {
		return 0, 0, nil
	}

	// caculate the total price
	totalPrice := caculateTotal(items, productMap)

	// reduce product's stock
	for _, item := range items {
		product := productMap[item.ProductID]
		product.Quantity -= item.Quantity

		// prevent race condition
		h.productStore.UpdateProduct(product)
	}

	// create the order & order items
	orderId, err := h.store.CreateOrder(types.Order{
		UserID: userID,
		Total: int(totalPrice),
		Status: "pending",
		Address: "some address :))",
	})
	if err != nil {
		return 0, 0, err
	}

	for _, item := range items {
		h.store.CreateOrderItem(types.OrderItem{
			OrderID: orderId,
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			Price: productMap[item.ProductID].Price,
		})
	}


	return orderId, totalPrice, nil
}

func checkIfCartIsInStock(items []types.CartItem, productMap map[int]types.Product) error {
	if len(items) == 0 {
		return fmt.Errorf("cart is empty")
	}

	for _, item := range items {
		product, ok := productMap[item.ProductID]
		if !ok {
			return fmt.Errorf("product %d is not available in store, please refresh your cart", item.ProductID)
		}

		if product.Quantity < item.Quantity {
			return fmt.Errorf("product %v is not available in quantity requested, please refresh your cart", product.Name)
		}

	}
	return nil

}

func caculateTotal(items []types.CartItem, productMap map[int]types.Product) float64 {
	var result float64
	for _, item := range items {
		product := productMap[item.ProductID]
		result += float64(item.Quantity) * product.Price
	}
	return result
}

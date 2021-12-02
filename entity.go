package inventory

import (
	"fmt"
	"log"

	domain "github.com/JoeQiao666/inventory/persistence"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type Inventory struct {
	inventory *domain.Inventory
}

func NewInventory(eventsourced.EntityID) eventsourced.EntityHandler {
	return &Inventory{
		inventory: new(domain.Inventory),
	}
}

func (i *Inventory) HandleEvent(ctx *eventsourced.Context, event interface{}) error {
	switch e := event.(type) {
	case *domain.ProductsAdded:
		return i.ProductsAdded(e)
	case *domain.ProductsReserved:
		return i.ProductsReserved(e)
	default:
		return nil
	}
}

func (i *Inventory) HandleCommand(ctx *eventsourced.Context, name string, cmd proto.Message) (proto.Message, error) {
	switch c := cmd.(type) {
	case *AddProducts:
		return i.Restock(ctx, c)
	case *ReserveProducts:
		return i.Reserve(ctx, c)
	case *QueryInventory:
		return i.GetInventory(ctx, c)
	default:
		return nil, nil
	}
}

func (i *Inventory) GetInventory(*eventsourced.Context, *QueryInventory) (*Products, error) {
	products := []*Product{}
	for _, v := range i.inventory.Products {
		products = append(products, &Product{
			ProductId: v.ProductId,
			Name:      v.Name,
			Quantity:  v.Quantity,
		})
	}
	return &Products{
		Products: products,
	}, nil
}

func (i *Inventory) Merge(product *domain.Product) {
	for _, v := range i.inventory.Products {
		if v.GetProductId() == product.GetProductId() {
			v.Quantity = v.GetQuantity() + product.GetQuantity()
			return
		}
	}
	i.inventory.Products = append(i.inventory.Products, product)
}

func (i *Inventory) Restock(ctx *eventsourced.Context, products *AddProducts) (*empty.Empty, error) {
	items := products.GetProducts()
	addedItems := []*domain.Product{}
	for _, v := range items {
		quantity := v.GetQuantity()
		if quantity < 0 {
			return nil, fmt.Errorf("can't add negative quantity %d of %s to inventory", quantity, v.GetName())
		}
		addedItems = append(addedItems, &domain.Product{
			ProductId: v.GetProductId(),
			Name:      v.GetName(),
			Quantity:  v.GetQuantity(),
		})
	}
	ctx.Emit(&domain.ProductsAdded{
		Products: addedItems,
	})
	return &empty.Empty{}, nil
}

func (i *Inventory) IsEnough(product *Product) error {
	for _, v := range i.inventory.Products {
		if v.GetProductId() == product.GetProductId() {
			if v.GetQuantity() < product.GetQuantity() {
				return fmt.Errorf("there are not enough %s left in the inventory", product.GetName())
			} else {
				return nil
			}
		}
	}
	return fmt.Errorf("there are no %s in the inventory", product.GetName())
}

func (i *Inventory) Reserve(ctx *eventsourced.Context, products *ReserveProducts) (*empty.Empty, error) {
	items := products.GetProducts()
	reservedItems := []*domain.Product{}
	for _, v := range items {
		err := i.IsEnough(v)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		reservedItems = append(reservedItems, &domain.Product{
			ProductId: v.GetProductId(),
			Name:      v.GetName(),
			Quantity:  v.GetQuantity(),
		})
	}
	ctx.Emit(&domain.ProductsReserved{
		Products: reservedItems,
	})
	return &empty.Empty{}, nil
}

func (i *Inventory) ProductsAdded(products *domain.ProductsAdded) error {
	for _, v := range products.GetProducts() {
		i.Merge(v)
	}
	return nil
}

func (i *Inventory) ProductsReserved(products *domain.ProductsReserved) error {
	for _, reserve := range products.GetProducts() {
		for idx, exist := range i.inventory.GetProducts() {
			if reserve.GetProductId() == exist.GetProductId() {
				exist.Quantity = exist.GetQuantity() - reserve.GetQuantity()
				if exist.Quantity == 0 {
					i.inventory.Products = append(i.inventory.Products[:idx], i.inventory.Products[idx+1:]...)
				}
			}
		}
	}
	return nil
}

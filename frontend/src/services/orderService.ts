import { CreateOrder, GetOrder, DeleteOrder, GetOrderWithPromotions, GetAllOrders } from "../../wailsjs/go/main/App";

export interface Order {
  id: number;
  customerName: string;
  totalPrice: number;
  itemCount: number;
  itemIDs: number[];
  promotions?: any[];
}

export const orderService = {
  create: async (customerName: string, itemIDs: number[]): Promise<number> => {
    return CreateOrder(customerName, itemIDs);
  },

  getById: async (id: number): Promise<Order> => {
    const result = await GetOrder(id);
    return {
      id: result.id,
      customerName: result.customer,
      totalPrice: result.totalPrice,
      itemCount: result.itemCount,
      itemIDs: result.itemIDs,
    };
  },

  getByIdWithPromotions: async (id: number): Promise<Order> => {
    const result = await GetOrderWithPromotions(id);
    return result as Order;
  },

  delete: async (id: number): Promise<void> => {
    return DeleteOrder(id);
  },

  getAll: async (): Promise<Order[]> => {
    const result = await GetAllOrders();
    return result as Order[];
  },
};

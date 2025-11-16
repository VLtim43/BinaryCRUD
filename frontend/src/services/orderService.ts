import { GetOrder, DeleteOrder, GetAllOrders } from "../../wailsjs/go/main/App";

export interface Order {
  id: number;
  customer: string;
  customerName?: string;
  totalPrice: number;
  itemCount: number;
  itemIDs: number[];
}

export const orderService = {
  getById: async (id: number): Promise<Order> => {
    const result = await GetOrder(id);
    return {
      id: result.id,
      customer: result.customer,
      totalPrice: result.totalPrice,
      itemCount: result.itemCount,
      itemIDs: result.itemIDs,
    };
  },

  delete: async (id: number): Promise<void> => {
    return DeleteOrder(id);
  },

  getAll: async (): Promise<Order[]> => {
    const result = await GetAllOrders();
    return result.map((item: any) => ({
      id: item.id,
      customer: item.customerName,
      totalPrice: item.totalPrice,
      itemCount: item.itemCount,
      itemIDs: item.itemIDs,
    }));
  },
};

import {
  CreateOrder,
  CreatePromotion,
  ApplyPromotionToOrder,
  GetOrderWithPromotions,
  GetOrderPromotions,
  RemovePromotionFromOrder,
} from "../../wailsjs/go/main/App";

export interface OrderWithPromotions {
  id: number;
  customerName: string;
  totalPrice: number;
  itemCount: number;
  itemIDs: number[];
  promotions: Array<{
    id: number;
    name: string;
    totalPrice: number;
    itemCount: number;
  }>;
}

export const orderPromotionService = {
  createOrder: async (customerName: string, itemIDs: number[]): Promise<number> => {
    return CreateOrder(customerName, itemIDs);
  },

  createPromotion: async (promotionName: string, itemIDs: number[]): Promise<number> => {
    return CreatePromotion(promotionName, itemIDs);
  },

  applyPromotionToOrder: async (orderID: number, promotionID: number): Promise<void> => {
    return ApplyPromotionToOrder(orderID, promotionID);
  },

  getOrderWithPromotions: async (orderID: number): Promise<OrderWithPromotions> => {
    const result = await GetOrderWithPromotions(orderID);
    return {
      id: result.id,
      customerName: result.customerName,
      totalPrice: result.totalPrice,
      itemCount: result.itemCount,
      itemIDs: result.itemIDs,
      promotions: result.promotions || [],
    };
  },

  getOrderPromotions: async (orderID: number): Promise<Array<any>> => {
    return GetOrderPromotions(orderID);
  },

  removePromotionFromOrder: async (orderID: number, promotionID: number): Promise<void> => {
    return RemovePromotionFromOrder(orderID, promotionID);
  },
};

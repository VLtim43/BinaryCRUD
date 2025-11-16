import { CreateOrder, CreatePromotion } from "../../wailsjs/go/main/App";

export const orderPromotionService = {
  createOrder: async (customerName: string, itemIDs: number[]): Promise<number> => {
    return CreateOrder(customerName, itemIDs);
  },

  createPromotion: async (promotionName: string, itemIDs: number[]): Promise<number> => {
    return CreatePromotion(promotionName, itemIDs);
  },
};

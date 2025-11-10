import {
  CreatePromotion,
  GetPromotion,
  DeletePromotion,
  GetAllPromotions,
  ApplyPromotionToOrder,
  RemovePromotionFromOrder,
} from "../../wailsjs/go/main/App";

export interface Promotion {
  id: number;
  name: string;
  totalPrice: number;
  itemCount: number;
  itemIDs?: number[];
}

export const promotionService = {
  create: async (name: string, itemIDs: number[]): Promise<number> => {
    return CreatePromotion(name, itemIDs);
  },

  getById: async (id: number): Promise<Promotion> => {
    const result = await GetPromotion(id);
    return {
      id: result.id,
      name: result.name,
      totalPrice: result.totalPrice,
      itemCount: result.itemCount,
      itemIDs: result.itemIDs,
    };
  },

  delete: async (id: number): Promise<void> => {
    return DeletePromotion(id);
  },

  getAll: async (): Promise<Promotion[]> => {
    const result = await GetAllPromotions();
    return result as Promotion[];
  },

  applyToOrder: async (orderId: number, promotionId: number): Promise<void> => {
    return ApplyPromotionToOrder(orderId, promotionId);
  },

  removeFromOrder: async (orderId: number, promotionId: number): Promise<void> => {
    return RemovePromotionFromOrder(orderId, promotionId);
  },
};

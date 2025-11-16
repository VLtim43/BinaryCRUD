import { GetPromotion, DeletePromotion, GetAllPromotions } from "../../wailsjs/go/main/App";

export interface Promotion {
  id: number;
  name: string;
  totalPrice: number;
  itemCount: number;
  itemIDs: number[];
}

export const promotionService = {
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
    return result.map((item: any) => ({
      id: item.id,
      name: item.name,
      totalPrice: item.totalPrice,
      itemCount: item.itemCount,
      itemIDs: item.itemIDs,
    }));
  },
};

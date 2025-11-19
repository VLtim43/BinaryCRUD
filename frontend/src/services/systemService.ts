import {
  DeleteAllFiles,
  PopulateInventory,
  PopulateItems,
  PopulatePromotions,
  PopulateOrders,
  GetIndexContents,
  GetOrderIndexContents,
  GetPromotionIndexContents
} from "../../wailsjs/go/main/App";

export const systemService = {
  deleteAllFiles: async (): Promise<void> => {
    return DeleteAllFiles();
  },

  populateInventory: async (): Promise<void> => {
    return PopulateInventory();
  },

  populateItems: async (): Promise<void> => {
    return PopulateItems();
  },

  populatePromotions: async (): Promise<void> => {
    return PopulatePromotions();
  },

  populateOrders: async (): Promise<void> => {
    return PopulateOrders();
  },

  getIndexContents: async (): Promise<any> => {
    return GetIndexContents();
  },

  getOrderIndexContents: async (): Promise<any> => {
    return GetOrderIndexContents();
  },

  getPromotionIndexContents: async (): Promise<any> => {
    return GetPromotionIndexContents();
  },
};

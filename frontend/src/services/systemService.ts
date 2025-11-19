import {
  DeleteAllFiles,
  PopulateInventory,
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

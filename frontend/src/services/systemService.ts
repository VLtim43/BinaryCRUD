import {
  DeleteAllFiles,
  PopulateInventory,
  GetIndexContents,
  GetOrderIndexContents,
  GetPromotionIndexContents,
  GetEncryptionEnabled,
  SetEncryptionEnabled,
  Compact
} from "../../wailsjs/go/main/App";

export interface CompactResult {
  itemsRemoved: number;
  ordersAffected: number;
  promotionsAffected: number;
  ordersRemoved: number;
  promotionsRemoved: number;
  orderPromotionsRemoved: number;
}

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

  getEncryptionEnabled: async (): Promise<boolean> => {
    return GetEncryptionEnabled();
  },

  setEncryptionEnabled: async (enabled: boolean): Promise<void> => {
    return SetEncryptionEnabled(enabled);
  },

  compact: async (): Promise<CompactResult> => {
    return Compact();
  },
};

import { DeleteAllFiles, PopulateInventory, GetIndexContents } from "../../wailsjs/go/main/App";

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
};

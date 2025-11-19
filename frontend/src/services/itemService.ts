import { AddItem, GetItem, DeleteItem, GetAllItems } from "../../wailsjs/go/main/App";

export interface Item {
  id: number;
  name: string;
  priceInCents: number;
  isDeleted?: boolean;
}

export const itemService = {
  create: async (name: string, priceInCents: number): Promise<number> => {
    return AddItem(name, priceInCents);
  },

  getById: async (id: number): Promise<Item> => {
    const result = await GetItem(id, true);
    return {
      id: result.id,
      name: result.name,
      priceInCents: result.priceInCents,
    };
  },

  delete: async (id: number): Promise<void> => {
    return DeleteItem(id);
  },

  getAll: async (): Promise<Item[]> => {
    const result = await GetAllItems();
    return result as Item[];
  },
};

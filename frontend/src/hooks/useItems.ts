import { useState } from "preact/hooks";
import { itemService, Item } from "../services/itemService";

export const useItems = () => {
  const [allItems, setAllItems] = useState<Item[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const loadAllItems = async () => {
    setIsLoading(true);
    try {
      const items = await itemService.getAll();
      setAllItems(items);
    } catch (err) {
      console.error("Error loading items:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const getActiveItems = () => {
    return allItems.filter((item) => !item.isDeleted);
  };

  return {
    allItems,
    isLoading,
    loadAllItems,
    getActiveItems,
  };
};

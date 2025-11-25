import { useState } from "preact/hooks";
import { itemService, Item } from "../services/itemService";
import { toast } from "../utils/toast";

export const useItems = () => {
  const [allItems, setAllItems] = useState<Item[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const loadAllItems = async () => {
    setIsLoading(true);
    try {
      const items = await itemService.getAll();
      setAllItems(items);
    } catch (err) {
      toast.error("Failed to load items");
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

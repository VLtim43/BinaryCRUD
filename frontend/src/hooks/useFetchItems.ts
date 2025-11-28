import { useState, useCallback } from "preact/hooks";
import { itemService, Item } from "../services/itemService";
import { toast } from "../utils/toast";

interface UseFetchItemsResult {
  items: Item[];
  loading: boolean;
  fetchItems: (itemIDs: number[]) => Promise<Item[] | null>;
  clearItems: () => void;
}

export function useFetchItems(): UseFetchItemsResult {
  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = useCallback(async (itemIDs: number[]): Promise<Item[] | null> => {
    if (!itemIDs || itemIDs.length === 0) {
      toast.warning("No items to display");
      return null;
    }

    setLoading(true);
    try {
      const fetchedItems = await Promise.all(
        itemIDs.map((id) => itemService.getById(id))
      );
      setItems(fetchedItems);
      return fetchedItems;
    } catch (err) {
      toast.error("Failed to fetch items");
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const clearItems = useCallback(() => {
    setItems([]);
  }, []);

  return { items, loading, fetchItems, clearItems };
}

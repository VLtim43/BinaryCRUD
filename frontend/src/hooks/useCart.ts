import { useState } from "preact/hooks";
import { Item } from "../services/itemService";
import { CartItem } from "../types/cart";
import { toast } from "../utils/toast";

interface UseCartOptions {
  onMessage: (msg: string) => void;
}

export const useCart = ({ onMessage }: UseCartOptions) => {
  const [cart, setCart] = useState<CartItem[]>([]);
  const [selectedItemId, setSelectedItemId] = useState("");

  const addItemToCart = (allItems: Item[]) => {
    if (!selectedItemId) {
      toast.warning("Please select an item");
      return;
    }

    const item = allItems.find((i) => i.id === parseInt(selectedItemId, 10));
    if (!item) {
      toast.warning("Item not found");
      return;
    }

    const existingItem = cart.find((c) => c.id === item.id);
    if (existingItem) {
      setCart(
        cart.map((c) =>
          c.id === item.id ? { ...c, quantity: c.quantity + 1 } : c
        )
      );
    } else {
      setCart([...cart, { ...item, quantity: 1 }]);
    }

    setSelectedItemId("");
  };

  const removeFromCart = (itemId: number) => {
    setCart(cart.filter((c) => c.id !== itemId));
  };

  const calculateTotal = () => {
    return cart.reduce(
      (sum, item) => sum + item.priceInCents * item.quantity,
      0
    );
  };

  const getTotalItemCount = () => {
    return cart.reduce((sum, item) => sum + item.quantity, 0);
  };

  const getItemIDs = (): number[] => {
    const itemIDs: number[] = [];
    cart.forEach((item) => {
      for (let i = 0; i < item.quantity; i++) {
        itemIDs.push(item.id);
      }
    });
    return itemIDs;
  };

  const clearCart = () => {
    setCart([]);
  };

  return {
    cart,
    selectedItemId,
    setSelectedItemId,
    addItemToCart,
    removeFromCart,
    calculateTotal,
    getTotalItemCount,
    getItemIDs,
    clearCart,
  };
};

import { useState } from "preact/hooks";

export interface CartItem {
  id: number;
  name: string;
  quantity: number;
  priceInCents: number;
}

export const useCart = () => {
  const [cart, setCart] = useState<CartItem[]>([]);

  const addItem = (item: { id: number; name: string; priceInCents: number }) => {
    const existingItem = cart.find((c) => c.id === item.id);
    if (existingItem) {
      setCart(
        cart.map((c) =>
          c.id === item.id ? { ...c, quantity: c.quantity + 1 } : c
        )
      );
    } else {
      setCart([
        ...cart,
        {
          id: item.id,
          name: item.name,
          quantity: 1,
          priceInCents: item.priceInCents,
        },
      ]);
    }
  };

  const removeItem = (itemId: number) => {
    setCart(cart.filter((c) => c.id !== itemId));
  };

  const clear = () => {
    setCart([]);
  };

  const total = cart.reduce((sum, item) => sum + item.priceInCents * item.quantity, 0);

  return {
    cart,
    addItem,
    removeItem,
    clear,
    total,
  };
};

import { h } from "preact";
import { formatPrice } from "../utils/formatters";

export interface Item {
  id: number;
  name: string;
  priceInCents: number;
}

interface ItemListProps {
  items: Item[];
}

export const ItemList = ({ items }: ItemListProps) => {
  return (
    <div className="cart-items" style={{ maxHeight: "400px", backgroundColor: "transparent", border: "none" }}>
      {items.map((item) => (
        <div key={item.id} className="cart-item">
          <div className="cart-item-info">
            <div className="cart-item-name">{item.name}</div>
            <div className="cart-item-id">ID: {item.id} | ${formatPrice(item.priceInCents)}</div>
          </div>
        </div>
      ))}
    </div>
  );
};

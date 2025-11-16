import { h } from "preact";
import { Button } from "./Button";
import { Badge } from "./Badge";
import "./CartItem.scss";

interface BaseCartItemProps {
  id: number;
  name: string;
  onRemove: (id: number) => void;
}

interface ItemCartItemProps extends BaseCartItemProps {
  type: "item";
  priceInCents: number;
  quantity: number;
}

interface PromoCartItemProps extends BaseCartItemProps {
  type: "promo";
  totalPrice: number;
  itemCount: number;
}

type CartItemProps = ItemCartItemProps | PromoCartItemProps;

export const CartItem = (props: CartItemProps) => {
  const { id, name, type, onRemove } = props;

  const formatPrice = (cents: number) => (cents / 100).toFixed(2);

  return (
    <div className={`cart-item-card cart-item-card-${type}`}>
      <div className="cart-item-info">
        <div className="cart-item-name">
          <Badge variant={type}>{type.toUpperCase()}</Badge>
          {name}
        </div>
        <div className="cart-item-details">
          {type === "item" ? (
            <>
              ID: {id} | ${formatPrice(props.priceInCents)} each | Total: $
              {formatPrice(props.priceInCents * props.quantity)}
            </>
          ) : (
            <>
              ID: {id} | {props.itemCount} items | ${formatPrice(props.totalPrice)}
            </>
          )}
        </div>
      </div>
      <div className="cart-item-actions">
        {type === "item" && <div className="cart-item-quantity">x{props.quantity}</div>}
        <Button variant="danger" size="small" onClick={() => onRemove(id)}>
          ×
        </Button>
      </div>
    </div>
  );
};

import { h } from "preact";
import { Button } from "./Button";
import { Input } from "./Input";
import { Select } from "./Select";
import { CreateFormLayout } from "./CreateFormLayout";
import { CartItem } from "./CartItem";

interface OrderCreateFormProps {
  customerName: string;
  onCustomerNameChange: (e: Event) => void;
  cart: Array<{ id: number; name: string; quantity: number; priceInCents: number }>;
  selectedPromotions: Array<{ id: number; name: string; totalPrice: number; itemCount: number }>;
  availableItems: Array<{ id: number; name: string; priceInCents: number }>;
  availablePromotions: Array<{ id: number; name: string; totalPrice: number; itemCount: number }>;
  selectedItemId: string;
  selectedPromotionId: string;
  onItemSelect: (e: Event) => void;
  onPromotionSelect: (e: Event) => void;
  onAddItem: () => void;
  onAddPromotion: () => void;
  onRemoveItem: (id: number) => void;
  onRemovePromotion: (id: number) => void;
  onSubmit: () => void;
}

export const OrderCreateForm = ({
  customerName,
  onCustomerNameChange,
  cart,
  selectedPromotions,
  availableItems,
  availablePromotions,
  selectedItemId,
  selectedPromotionId,
  onItemSelect,
  onPromotionSelect,
  onAddItem,
  onAddPromotion,
  onRemoveItem,
  onRemovePromotion,
  onSubmit,
}: OrderCreateFormProps) => {
  const orderTotal =
    (cart.reduce((sum, item) => sum + item.priceInCents * item.quantity, 0) +
      selectedPromotions.reduce((sum, promo) => sum + promo.totalPrice, 0)) /
    100;

  const itemOptions = availableItems
    .sort((a, b) => a.id - b.id)
    .map((item) => ({
      value: item.id,
      label: `[${item.id}] ${item.name} - $${(item.priceInCents / 100).toFixed(2)}`,
    }));

  const promotionOptions = availablePromotions
    .filter((p) => !selectedPromotions.some((sp) => sp.id === p.id))
    .sort((a, b) => a.id - b.id)
    .map((promo) => ({
      value: promo.id,
      label: `[${promo.id}] ${promo.name} - $${(promo.totalPrice / 100).toFixed(2)}`,
    }));

  return (
    <CreateFormLayout
      title="Create Order"
      submitDisabled={!customerName || (cart.length === 0 && selectedPromotions.length === 0)}
      onSubmit={onSubmit}
      headerInputs={
        <Input
          value={customerName}
          onChange={onCustomerNameChange}
          placeholder="Customer Name"
          className="create-form-input-full"
        />
      }
      totalLabel="Order Total"
      totalAmount={orderTotal.toFixed(2)}
      contentLabel="Order Contents"
      contentEmpty={cart.length === 0 && selectedPromotions.length === 0}
      emptyMessage="No items or promotions added"
      footer={
        <div className="create-form-footer-row">
          <Select
            value={selectedItemId}
            onChange={onItemSelect}
            options={itemOptions}
            placeholder="Select an item..."
            className="flex-1"
          />
          <Button onClick={onAddItem} className="whitespace-nowrap">
            Add Item
          </Button>

          <Select
            value={selectedPromotionId}
            onChange={onPromotionSelect}
            options={promotionOptions}
            placeholder="Select a promotion..."
            className="flex-1"
          />
          <Button onClick={onAddPromotion} className="whitespace-nowrap">
            Add Promo
          </Button>
        </div>
      }
    >
      {cart.map((item) => (
        <CartItem
          key={`item-${item.id}`}
          type="item"
          id={item.id}
          name={item.name}
          priceInCents={item.priceInCents}
          quantity={item.quantity}
          onRemove={onRemoveItem}
        />
      ))}
      {selectedPromotions.map((promo) => (
        <CartItem
          key={`promo-${promo.id}`}
          type="promo"
          id={promo.id}
          name={promo.name}
          totalPrice={promo.totalPrice}
          itemCount={promo.itemCount}
          onRemove={onRemovePromotion}
        />
      ))}
    </CreateFormLayout>
  );
};

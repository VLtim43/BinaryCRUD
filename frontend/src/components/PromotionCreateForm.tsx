import { h } from "preact";
import { Button } from "./Button";
import { Input } from "./Input";
import { Select } from "./Select";
import { CreateFormLayout } from "./CreateFormLayout";
import { CartItem } from "./CartItem";

interface PromotionCreateFormProps {
  promotionName: string;
  onPromotionNameChange: (e: Event) => void;
  cart: Array<{ id: number; name: string; quantity: number; priceInCents: number }>;
  availableItems: Array<{ id: number; name: string; priceInCents: number }>;
  selectedItemId: string;
  onItemSelect: (e: Event) => void;
  onAddItem: () => void;
  onRemoveItem: (id: number) => void;
  onSubmit: () => void;
}

export const PromotionCreateForm = ({
  promotionName,
  onPromotionNameChange,
  cart,
  availableItems,
  selectedItemId,
  onItemSelect,
  onAddItem,
  onRemoveItem,
  onSubmit,
}: PromotionCreateFormProps) => {
  const promotionTotal =
    cart.reduce((sum, item) => sum + item.priceInCents * item.quantity, 0) / 100;

  const itemOptions = availableItems
    .sort((a, b) => a.id - b.id)
    .map((item) => ({
      value: item.id,
      label: `[${item.id}] ${item.name} - $${(item.priceInCents / 100).toFixed(2)}`,
    }));

  return (
    <CreateFormLayout
      title="Create Promotion"
      submitDisabled={!promotionName || cart.length === 0}
      onSubmit={onSubmit}
      headerInputs={
        <Input
          value={promotionName}
          onChange={onPromotionNameChange}
          placeholder="Promotion Name"
          className="create-form-input-full"
        />
      }
      totalLabel="Promotion Total"
      totalAmount={promotionTotal.toFixed(2)}
      contentLabel="Promotion Items"
      contentEmpty={cart.length === 0}
      emptyMessage="No items in promotion"
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
        </div>
      }
    >
      {cart.map((item) => (
        <CartItem
          key={item.id}
          type="item"
          id={item.id}
          name={item.name}
          priceInCents={item.priceInCents}
          quantity={item.quantity}
          onRemove={onRemoveItem}
        />
      ))}
    </CreateFormLayout>
  );
};

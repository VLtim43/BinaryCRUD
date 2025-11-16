import { h } from "preact";
import { Button } from "./Button";
import { Input } from "./Input";
import { Select } from "./Select";
import { CartItem } from "./CartItem";
import "./CreateFormLayout.scss";

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
    <div className="create-form-container">
      <div className="create-form">
        {/* Fixed Header */}
        <div className="create-form-header">
          <div className="create-form-title-row">
            <h3>Create Promotion</h3>
            <button
              className="btn btn-primary"
              onClick={onSubmit}
              disabled={!promotionName || cart.length === 0}
            >
              Submit
            </button>
          </div>
          <Input
            value={promotionName}
            onChange={onPromotionNameChange}
            placeholder="Promotion Name"
            className="create-form-input-full"
          />
          <div className="cart-total">Promotion Total: ${promotionTotal.toFixed(2)}</div>
        </div>

        {/* Scrollable Content */}
        <div className="create-form-content">
          <h4 className="create-form-content-label">Promotion Items</h4>
          <div className="create-form-items-container">
            {cart.length === 0 ? (
              <div className="empty-state">No items in promotion</div>
            ) : (
              <div className="create-form-items-list">
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
              </div>
            )}
          </div>
        </div>

        {/* Fixed Footer */}
        <div className="create-form-footer">
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
        </div>
      </div>
    </div>
  );
};

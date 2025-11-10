import { h } from "preact";
import { Button } from "./Button";
import { Input } from "./Input";
import { Select } from "./Select";

interface OrderCreateFormProps {
  customerName: string;
  onCustomerNameChange: (e: any) => void;
  cart: Array<{ id: number; name: string; quantity: number; priceInCents: number }>;
  selectedPromotions: Array<{ id: number; name: string; totalPrice: number; itemCount: number }>;
  availableItems: Array<{ id: number; name: string; priceInCents: number }>;
  availablePromotions: Array<{ id: number; name: string; totalPrice: number; itemCount: number }>;
  selectedItemId: string;
  selectedPromotionId: string;
  onItemSelect: (e: any) => void;
  onPromotionSelect: (e: any) => void;
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
    <div
      id="order-section"
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "flex-start",
        paddingTop: "20px",
        width: "100%",
      }}
    >
      <div
        style={{
          width: "600px",
          maxWidth: "90vw",
          height: "275px",
          display: "flex",
          flexDirection: "column",
          backgroundColor: "rgba(255, 255, 255, 0.05)",
          borderRadius: "8px",
          border: "1px solid rgba(255, 255, 255, 0.1)",
          overflow: "hidden",
        }}
      >
        {/* Fixed Header */}
        <div
          style={{
            padding: "20px",
            borderBottom: "1px solid rgba(255, 255, 255, 0.1)",
            flexShrink: 0,
          }}
        >
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "15px" }}>
            <h3 style={{ margin: 0, color: "#fff" }}>Create Order</h3>
            <Button
              onClick={onSubmit}
              disabled={!customerName || (cart.length === 0 && selectedPromotions.length === 0)}
            >
              Submit
            </Button>
          </div>
          <Input
            value={customerName}
            onChange={onCustomerNameChange}
            placeholder="Customer Name"
            style={{ width: "100%", marginBottom: "10px", padding: "8px", boxSizing: "border-box" }}
          />
          <div className="cart-total" style={{ marginBottom: 0 }}>
            Order Total: ${orderTotal.toFixed(2)}
          </div>
        </div>

        {/* Scrollable Content */}
        <div
          style={{
            flex: 1,
            overflow: "hidden",
            padding: "20px",
            display: "flex",
            flexDirection: "column",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0", color: "#fff", fontSize: "0.9em" }}>Order Contents</h4>
          <div
            style={{
              flex: 1,
              overflowY: "auto",
              border: "1px solid rgba(100, 150, 255, 0.2)",
              borderRadius: "4px",
              padding: "8px",
            }}
          >
            {cart.length === 0 && selectedPromotions.length === 0 ? (
              <div style={{ padding: "20px", textAlign: "center", color: "#aaa", fontStyle: "italic" }}>
                No items or promotions added
              </div>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
                {cart.map((item) => (
                  <div
                    key={`item-${item.id}`}
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      padding: "10px",
                      backgroundColor: "rgba(100, 150, 255, 0.05)",
                      border: "1px solid rgba(100, 150, 255, 0.2)",
                      borderRadius: "4px",
                    }}
                  >
                    <div>
                      <div style={{ color: "#fff" }}>
                        <span
                          style={{
                            display: "inline-block",
                            padding: "2px 6px",
                            backgroundColor: "rgba(100, 150, 255, 0.3)",
                            borderRadius: "3px",
                            fontSize: "0.75em",
                            fontWeight: "bold",
                            marginRight: "8px",
                          }}
                        >
                          ITEM
                        </span>
                        {item.name}
                      </div>
                      <div style={{ color: "#aaa", fontSize: "0.85em" }}>
                        ID: {item.id} | ${(item.priceInCents / 100).toFixed(2)} each | Total: $
                        {((item.priceInCents * item.quantity) / 100).toFixed(2)}
                      </div>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                      <div style={{ color: "#fff", fontWeight: "bold" }}>x{item.quantity}</div>
                      <Button variant="danger" size="small" onClick={() => onRemoveItem(item.id)}>
                        ×
                      </Button>
                    </div>
                  </div>
                ))}
                {selectedPromotions.map((promo) => (
                  <div
                    key={`promo-${promo.id}`}
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      padding: "10px",
                      backgroundColor: "rgba(100, 200, 100, 0.05)",
                      border: "1px solid rgba(100, 200, 100, 0.2)",
                      borderRadius: "4px",
                    }}
                  >
                    <div>
                      <div style={{ color: "#fff" }}>
                        <span
                          style={{
                            display: "inline-block",
                            padding: "2px 6px",
                            backgroundColor: "rgba(100, 200, 100, 0.3)",
                            borderRadius: "3px",
                            fontSize: "0.75em",
                            fontWeight: "bold",
                            marginRight: "8px",
                          }}
                        >
                          PROMO
                        </span>
                        {promo.name}
                      </div>
                      <div style={{ color: "#aaa", fontSize: "0.85em" }}>
                        ID: {promo.id} | {promo.itemCount} items | ${(promo.totalPrice / 100).toFixed(2)}
                      </div>
                    </div>
                    <Button variant="danger" size="small" onClick={() => onRemovePromotion(promo.id)}>
                      ×
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Fixed Footer */}
        <div
          style={{
            padding: "15px 20px",
            borderTop: "1px solid rgba(255, 255, 255, 0.1)",
            flexShrink: 0,
            backgroundColor: "rgba(0, 0, 0, 0.2)",
          }}
        >
          <div style={{ display: "flex", gap: "8px", flexDirection: "row" }}>
            <Select
              value={selectedItemId}
              onChange={onItemSelect}
              options={itemOptions}
              placeholder="Select an item..."
              style={{ flex: 1, minWidth: 0 }}
            />
            <Button onClick={onAddItem} style={{ whiteSpace: "nowrap" }}>
              Add Item
            </Button>

            <Select
              value={selectedPromotionId}
              onChange={onPromotionSelect}
              options={promotionOptions}
              placeholder="Select a promotion..."
              style={{ flex: 1, minWidth: 0 }}
            />
            <Button onClick={onAddPromotion} style={{ whiteSpace: "nowrap" }}>
              Add Promo
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};

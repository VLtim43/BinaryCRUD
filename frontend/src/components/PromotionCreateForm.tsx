import { h } from "preact";
import { Button } from "./Button";
import { Input } from "./Input";
import { Select } from "./Select";

interface PromotionCreateFormProps {
  promotionName: string;
  onPromotionNameChange: (e: any) => void;
  cart: Array<{ id: number; name: string; quantity: number; priceInCents: number }>;
  availableItems: Array<{ id: number; name: string; priceInCents: number }>;
  selectedItemId: string;
  onItemSelect: (e: any) => void;
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
    <div
      id="promotion-section"
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
            <h3 style={{ margin: 0, color: "#fff" }}>Create Promotion</h3>
            <Button onClick={onSubmit} disabled={!promotionName || cart.length === 0}>
              Submit
            </Button>
          </div>
          <Input
            value={promotionName}
            onChange={onPromotionNameChange}
            placeholder="Promotion Name"
            style={{ width: "100%", marginBottom: "10px", padding: "8px", boxSizing: "border-box" }}
          />
          <div className="cart-total" style={{ marginBottom: 0 }}>
            Promotion Total: ${promotionTotal.toFixed(2)}
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
          <h4 style={{ margin: "0 0 10px 0", color: "#fff", fontSize: "0.9em" }}>Promotion Items</h4>
          <div
            style={{
              flex: 1,
              overflowY: "auto",
              border: "1px solid rgba(100, 150, 255, 0.2)",
              borderRadius: "4px",
              padding: "8px",
            }}
          >
            {cart.length === 0 ? (
              <div style={{ padding: "20px", textAlign: "center", color: "#aaa", fontStyle: "italic" }}>
                No items in promotion
              </div>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
                {cart.map((item) => (
                  <div
                    key={item.id}
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      padding: "10px",
                      backgroundColor: "rgba(255, 255, 255, 0.05)",
                      border: "1px solid rgba(255, 255, 255, 0.2)",
                      borderRadius: "4px",
                    }}
                  >
                    <div>
                      <div style={{ color: "#fff" }}>{item.name}</div>
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
          <div style={{ display: "flex", gap: "8px" }}>
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
          </div>
        </div>
      </div>
    </div>
  );
};

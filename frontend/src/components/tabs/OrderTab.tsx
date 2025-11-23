import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { Select } from "../Select";
import { Modal } from "../Modal";
import { ItemList } from "../ItemList";
import { SubTabs } from "../SubTabs";
import { DeleteForm } from "../DeleteForm";
import { orderService } from "../../services/orderService";
import { itemService, Item } from "../../services/itemService";
import { promotionService, Promotion } from "../../services/promotionService";
import {
  orderPromotionService,
  OrderWithPromotions,
} from "../../services/orderPromotionService";
import {
  formatPrice,
  isValidId,
  createIdInputHandler,
  createInputHandler,
  createSelectHandler,
  formatError,
  PROMO_CARD_STYLE,
  CRUD_TABS,
} from "../../utils/formatters";
import { useCart } from "../../hooks/useCart";
import { useItems } from "../../hooks/useItems";

interface OrderTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const OrderTab = ({ onMessage, onRefreshLogs }: OrderTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [customerName, setCustomerName] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundOrder, setFoundOrder] = useState<OrderWithPromotions | null>(null);
  const [isItemModalOpen, setIsItemModalOpen] = useState(false);
  const [isPromoModalOpen, setIsPromoModalOpen] = useState(false);
  const [items, setItems] = useState<Item[]>([]);
  const [promoItems, setPromoItems] = useState<Item[]>([]);
  const [selectedPromoForView, setSelectedPromoForView] = useState<{
    id: number;
    name: string;
  } | null>(null);
  const [allPromotions, setAllPromotions] = useState<Promotion[]>([]);
  const [selectedPromotionId, setSelectedPromotionId] = useState("");
  const [selectedPromotions, setSelectedPromotions] = useState<Promotion[]>([]);

  const { allItems, loadAllItems, getActiveItems } = useItems();
  const {
    cart,
    selectedItemId,
    setSelectedItemId,
    addItemToCart,
    removeFromCart,
    calculateTotal,
    getTotalItemCount,
    getItemIDs,
    clearCart,
  } = useCart({ onMessage });

  useEffect(() => {
    if (subTab === "create") {
      loadAllItems();
      loadAllPromotions();
    }
  }, [subTab]);

  const loadAllPromotions = async () => {
    try {
      const promotions = await promotionService.getAll();
      setAllPromotions(promotions);
    } catch (err) {
      console.error("Error loading promotions:", err);
    }
  };

  const handleRead = async () => {
    if (!isValidId(recordId)) {
      onMessage("Error: Please enter a valid record ID");
      setFoundOrder(null);
      return;
    }

    try {
      const order = await orderPromotionService.getOrderWithPromotions(
        parseInt(recordId, 10)
      );
      setFoundOrder(order);
      const promoCount = order.promotions?.length || 0;
      onMessage(
        `Found Order #${order.id}: ${order.customerName} - $${formatPrice(
          order.totalPrice
        )} (${order.itemCount} items${
          promoCount > 0 ? `, ${promoCount} promotions` : ""
        })`
      );
      onRefreshLogs();
    } catch (err) {
      setFoundOrder(null);
      onMessage(`Error: ${formatError(err)}`);
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      onMessage("Error: Please enter a valid record ID");
      return;
    }

    try {
      await orderService.delete(parseInt(deleteId, 10));
      onMessage(`Successfully deleted order with ID ${deleteId}`);
      setDeleteId("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${formatError(err)}`);
    }
  };

  const handleShowItems = async () => {
    if (!foundOrder || !foundOrder.itemIDs || foundOrder.itemIDs.length === 0) {
      onMessage("No items to display");
      return;
    }

    try {
      const fetchedItems = await Promise.all(
        foundOrder.itemIDs.map((id) => itemService.getById(id))
      );
      setItems(fetchedItems);
      setIsItemModalOpen(true);
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error fetching items: ${formatError(err)}`);
    }
  };

  const handleShowPromotionItems = async (
    promotionId: number,
    promotionName: string
  ) => {
    try {
      const promotion = await promotionService.getById(promotionId);
      if (!promotion.itemIDs || promotion.itemIDs.length === 0) {
        onMessage("No items in this promotion");
        return;
      }

      const fetchedItems = await Promise.all(
        promotion.itemIDs.map((id) => itemService.getById(id))
      );
      setPromoItems(fetchedItems);
      setSelectedPromoForView({ id: promotionId, name: promotionName });
      setIsPromoModalOpen(true);
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error fetching promotion items: ${formatError(err)}`);
    }
  };

  const handleAddPromotion = () => {
    if (!selectedPromotionId) {
      onMessage("Please select a promotion");
      return;
    }

    if (selectedPromotions.length >= 1) {
      onMessage("Only one promotion can be added per order");
      return;
    }

    const promotion = allPromotions.find(
      (p) => p.id === parseInt(selectedPromotionId, 10)
    );
    if (!promotion) {
      onMessage("Promotion not found");
      return;
    }

    setSelectedPromotions([promotion]);
    setSelectedPromotionId("");
  };

  const handleRemovePromotion = (promotionId: number) => {
    setSelectedPromotions(
      selectedPromotions.filter((p) => p.id !== promotionId)
    );
  };

  const calculateOrderTotal = () => {
    const itemsTotal = calculateTotal();
    const promotionsTotal = selectedPromotions.reduce(
      (sum, promo) => sum + promo.totalPrice,
      0
    );
    return itemsTotal + promotionsTotal;
  };

  const handleCreateOrder = async () => {
    if (!customerName || customerName.trim().length === 0) {
      onMessage("Error: Please enter a customer name");
      return;
    }

    if (cart.length === 0) {
      onMessage("Error: Please add at least one item to the order");
      return;
    }

    try {
      const itemIDs = getItemIDs();
      const orderId = await orderPromotionService.createOrder(
        customerName,
        itemIDs
      );

      for (const promotion of selectedPromotions) {
        await orderPromotionService.applyPromotionToOrder(
          orderId,
          promotion.id
        );
      }

      const promoCount = selectedPromotions.length;
      onMessage(
        `Order #${orderId} created successfully for ${customerName} ($${formatPrice(
          calculateOrderTotal()
        )})${promoCount > 0 ? ` with ${promoCount} promotion(s)` : ""}`
      );
      setCustomerName("");
      clearCart();
      setSelectedPromotions([]);
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${formatError(err)}`);
    }
  };

  return (
    <>
      <SubTabs
        tabs={[...CRUD_TABS]}
        activeTab={subTab}
        onTabChange={(tab) => setSubTab(tab as typeof subTab)}
      />

      {subTab === "create" && (
        <>
          <div className="cart-container">
            <div
              className="cart-header"
              style={{ display: "flex", flexDirection: "column", gap: "10px" }}
            >
              <div
                style={{ display: "flex", gap: "10px", alignItems: "center" }}
              >
                <Select
                  value={selectedItemId}
                  onChange={createSelectHandler(setSelectedItemId)}
                  options={getActiveItems().map((item) => ({
                    value: item.id,
                    label: `${item.name} - $${formatPrice(item.priceInCents)}`,
                  }))}
                  placeholder="Select an item..."
                  className="cart-select"
                />
                <Button onClick={() => addItemToCart(allItems)}>Add Item</Button>
              </div>
              <div
                style={{ display: "flex", gap: "10px", alignItems: "center" }}
              >
                <Select
                  value={selectedPromotionId}
                  onChange={createSelectHandler(setSelectedPromotionId)}
                  options={allPromotions.map((promo) => ({
                    value: promo.id,
                    label: `${promo.name} - $${formatPrice(promo.totalPrice)}`,
                  }))}
                  placeholder={
                    selectedPromotions.length >= 1
                      ? "Max 1 promotion per order"
                      : "Select a promotion..."
                  }
                  className="cart-select"
                  style={
                    selectedPromotions.length >= 1
                      ? { opacity: 0.5, pointerEvents: "none" }
                      : {}
                  }
                />
                <Button
                  onClick={handleAddPromotion}
                  style={
                    selectedPromotions.length >= 1
                      ? { opacity: 0.5, pointerEvents: "none" }
                      : {}
                  }
                >
                  Add Promotion
                </Button>
              </div>
            </div>

            <div className="cart-total">
              Total: ${formatPrice(calculateOrderTotal())} (
              {getTotalItemCount()} items
              {selectedPromotions.length > 0 &&
                `, ${selectedPromotions.length} promotions`}
              )
            </div>

            <div className="cart-items">
              {cart.length === 0 && selectedPromotions.length === 0 ? (
                <div className="cart-empty">
                  No items or promotions added yet
                </div>
              ) : (
                <>
                  {cart.map((item) => (
                    <div key={`item-${item.id}`} className="cart-item">
                      <div className="cart-item-info">
                        <div className="cart-item-name">{item.name}</div>
                        <div className="cart-item-id">
                          ID: {item.id} | ${formatPrice(item.priceInCents)} each
                        </div>
                      </div>
                      <div className="cart-item-controls">
                        <div className="cart-item-quantity">
                          x{item.quantity}
                        </div>
                        <Button
                          size="small"
                          variant="danger"
                          onClick={() => removeFromCart(item.id)}
                        >
                          ×
                        </Button>
                      </div>
                    </div>
                  ))}
                  {selectedPromotions.map((promo) => (
                    <div
                      key={`promo-${promo.id}`}
                      className="cart-item"
                      style={PROMO_CARD_STYLE}
                    >
                      <div className="cart-item-info">
                        <div className="cart-item-name">
                          [PROMO] {promo.name}
                        </div>
                        <div className="cart-item-id">
                          ID: {promo.id} | ${formatPrice(promo.totalPrice)} |{" "}
                          {promo.itemCount} items
                        </div>
                      </div>
                      <div className="cart-item-controls">
                        <Button
                          size="small"
                          variant="danger"
                          onClick={() => handleRemovePromotion(promo.id)}
                        >
                          ×
                        </Button>
                      </div>
                    </div>
                  ))}
                </>
              )}
            </div>

            <div className="cart-footer">
              <div
                className="input-box"
                style={{ height: "35px", margin: 0, flex: 1 }}
              >
                <Input
                  id="customer-name"
                  placeholder="Customer Name"
                  value={customerName}
                  onChange={createInputHandler(setCustomerName)}
                />
                <Button variant="primary" onClick={handleCreateOrder}>
                  Create Order
                </Button>
              </div>
            </div>
          </div>
        </>
      )}

      {subTab === "read" && (
        <>
          <div className="input-box">
            <Input
              id="record-id"
              placeholder="Enter Record ID"
              value={recordId}
              onChange={createIdInputHandler(setRecordId)}
            />
            <Button onClick={handleRead}>Get Record</Button>
          </div>

          {foundOrder && (
            <div className="details-card">
              <h3>Order Details</h3>
              <div className="details-content">
                <div className="details-row">
                  <span className="details-label">ID:</span>
                  <span className="details-value">{foundOrder.id}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Customer:</span>
                  <span className="details-value">
                    {foundOrder.customerName}
                  </span>
                </div>
                <div className="details-row">
                  <span className="details-label">Total Price:</span>
                  <span className="details-value">
                    ${formatPrice(foundOrder.totalPrice)}
                  </span>
                </div>
                <div className="details-row">
                  <span className="details-label">Item Count:</span>
                  <span className="details-value">{foundOrder.itemCount}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Item IDs:</span>
                  <span
                    className="details-value"
                    onClick={handleShowItems}
                    style={{ cursor: "pointer" }}
                  >
                    {foundOrder.itemIDs.join(", ")}
                  </span>
                </div>
                {foundOrder.promotions && foundOrder.promotions.length > 0 && (
                  <>
                    {foundOrder.promotions.map((promo) => (
                      <div
                        key={promo.id}
                        className="details-row"
                        style={{
                          ...PROMO_CARD_STYLE,
                          backgroundColor: "rgba(100, 200, 100, 0.1)",
                          cursor: "pointer",
                        }}
                        onClick={() =>
                          handleShowPromotionItems(promo.id, promo.name)
                        }
                      >
                        <span className="details-label">
                          [PROMO] {promo.name}:
                        </span>
                        <span className="details-value">
                          ${formatPrice(promo.totalPrice)} ({promo.itemCount}{" "}
                          items)
                        </span>
                      </div>
                    ))}
                  </>
                )}
              </div>
            </div>
          )}
        </>
      )}

      {subTab === "delete" && (
        <DeleteForm
          deleteId={deleteId}
          setDeleteId={setDeleteId}
          onDelete={handleDelete}
          entityName="Record"
        />
      )}

      <Modal
        isOpen={isItemModalOpen}
        onClose={() => setIsItemModalOpen(false)}
        title="Order Items"
      >
        <ItemList items={items} />
      </Modal>

      <Modal
        isOpen={isPromoModalOpen}
        onClose={() => setIsPromoModalOpen(false)}
        title={
          selectedPromoForView
            ? `Promotion: ${selectedPromoForView.name}`
            : "Promotion Items"
        }
      >
        <ItemList items={promoItems} />
      </Modal>
    </>
  );
};

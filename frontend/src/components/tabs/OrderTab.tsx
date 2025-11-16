import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { Select } from "../Select";
import { DataTable } from "../DataTable";
import { Modal } from "../Modal";
import { orderService, Order } from "../../services/orderService";
import { itemService, Item } from "../../services/itemService";
import { orderPromotionService } from "../../services/orderPromotionService";
import { formatPrice, isValidId, createIdInputHandler } from "../../utils/formatters";

interface OrderTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

interface CartItem {
  id: number;
  name: string;
  priceInCents: number;
  quantity: number;
}

export const OrderTab = ({ onMessage, onRefreshLogs }: OrderTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [customerName, setCustomerName] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundOrder, setFoundOrder] = useState<Order | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [items, setItems] = useState<Item[]>([]);
  const [allItems, setAllItems] = useState<Item[]>([]);
  const [selectedItemId, setSelectedItemId] = useState("");
  const [cart, setCart] = useState<CartItem[]>([]);

  useEffect(() => {
    if (subTab === "create") {
      loadAllItems();
    }
  }, [subTab]);

  const loadAllItems = async () => {
    try {
      const items = await itemService.getAll();
      setAllItems(items);
    } catch (err) {
      console.error("Error loading items:", err);
    }
  };

  const handleRead = async () => {
    if (!isValidId(recordId)) {
      onMessage("Error: Please enter a valid record ID");
      setFoundOrder(null);
      return;
    }

    try {
      const order = await orderService.getById(parseInt(recordId, 10));
      setFoundOrder(order);
      onMessage(`Found Order #${order.id}: ${order.customer} - $${formatPrice(order.totalPrice)} (${order.itemCount} items)`);
      onRefreshLogs();
    } catch (err) {
      setFoundOrder(null);
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
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
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
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
      setIsModalOpen(true);
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error fetching items: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleAddItemToCart = () => {
    if (!selectedItemId) {
      onMessage("Please select an item");
      return;
    }

    const item = allItems.find((i) => i.id === parseInt(selectedItemId, 10));
    if (!item) {
      onMessage("Item not found");
      return;
    }

    const existingItem = cart.find((c) => c.id === item.id);
    if (existingItem) {
      setCart(cart.map((c) => (c.id === item.id ? { ...c, quantity: c.quantity + 1 } : c)));
    } else {
      setCart([...cart, { ...item, quantity: 1 }]);
    }

    setSelectedItemId("");
  };

  const handleRemoveFromCart = (itemId: number) => {
    setCart(cart.filter((c) => c.id !== itemId));
  };

  const calculateTotal = () => {
    return cart.reduce((sum, item) => sum + item.priceInCents * item.quantity, 0);
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
      const itemIDs: number[] = [];
      cart.forEach((item) => {
        for (let i = 0; i < item.quantity; i++) {
          itemIDs.push(item.id);
        }
      });

      const orderId = await orderPromotionService.createOrder(customerName, itemIDs);
      onMessage(`Order #${orderId} created successfully for ${customerName} ($${formatPrice(calculateTotal())})`);
      setCustomerName("");
      setCart([]);
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  return (
    <>
      <div className="sub_tabs">
        <Button className={`tab ${subTab === "create" ? "active" : ""}`} onClick={() => setSubTab("create")}>
          Create
        </Button>
        <Button className={`tab ${subTab === "read" ? "active" : ""}`} onClick={() => setSubTab("read")}>
          Read
        </Button>
        <Button className={`tab ${subTab === "delete" ? "active" : ""}`} onClick={() => setSubTab("delete")}>
          Delete
        </Button>
      </div>

      {subTab === "create" && (
        <>
          <div className="cart-container">
            <div className="cart-header">
              <Select
                value={selectedItemId}
                onChange={(e: Event) => {
                  const target = e.target as HTMLSelectElement;
                  setSelectedItemId(target.value);
                }}
                options={allItems.map((item) => ({
                  value: item.id,
                  label: `${item.name} - $${formatPrice(item.priceInCents)}`,
                }))}
                placeholder="Select an item..."
                className="cart-select"
              />
              <Button onClick={handleAddItemToCart}>Add Item</Button>
            </div>

            <div className="cart-total">
              Total: ${formatPrice(calculateTotal())} ({cart.reduce((sum, item) => sum + item.quantity, 0)} items)
            </div>

            <div className="cart-items">
              {cart.length === 0 ? (
                <div className="cart-empty">No items added yet</div>
              ) : (
                cart.map((item) => (
                  <div key={item.id} className="cart-item">
                    <div className="cart-item-info">
                      <div className="cart-item-name">{item.name}</div>
                      <div className="cart-item-id">ID: {item.id} | ${formatPrice(item.priceInCents)} each</div>
                    </div>
                    <div className="cart-item-controls">
                      <div className="cart-item-quantity">x{item.quantity}</div>
                      <Button size="small" variant="danger" onClick={() => handleRemoveFromCart(item.id)}>
                        ×
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </div>

            <div className="cart-footer">
              <div className="input-box" style={{ height: "35px", margin: 0, flex: 1 }}>
                <Input
                  id="customer-name"
                  placeholder="Customer Name"
                  value={customerName}
                  onChange={(e: Event) => {
                    const target = e.target as HTMLInputElement;
                    setCustomerName(target.value);
                  }}
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
                  <span className="details-value">{foundOrder.customer}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Total Price:</span>
                  <span className="details-value">${formatPrice(foundOrder.totalPrice)}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Item Count:</span>
                  <span className="details-value">{foundOrder.itemCount}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Item IDs:</span>
                  <span className="details-value clickable-item-ids" onClick={handleShowItems}>
                    {foundOrder.itemIDs.join(", ")}
                  </span>
                </div>
              </div>
            </div>
          )}
        </>
      )}

      {subTab === "delete" && (
        <div className="input-box">
          <Input
            id="delete-record-id"
            placeholder="Enter Record ID"
            value={deleteId}
            onChange={createIdInputHandler(setDeleteId)}
          />
          <Button variant="danger" onClick={handleDelete}>
            Delete Record
          </Button>
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} title="Order Items">
        <div className="item-details-grid">
          {items.map((item) => (
            <div key={item.id} className="item-details-card">
              <h4>{item.name}</h4>
              <div className="item-detail-row">
                <span className="item-detail-label">ID:</span>
                <span className="item-detail-value">{item.id}</span>
              </div>
              <div className="item-detail-row">
                <span className="item-detail-label">Price:</span>
                <span className="item-detail-value">${formatPrice(item.priceInCents)}</span>
              </div>
            </div>
          ))}
        </div>
      </Modal>
    </>
  );
};

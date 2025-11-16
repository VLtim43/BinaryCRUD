import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { Select } from "../Select";
import { DataTable } from "../DataTable";
import { Modal } from "../Modal";
import { ItemList } from "../ItemList";
import { promotionService, Promotion } from "../../services/promotionService";
import { itemService, Item } from "../../services/itemService";
import { orderPromotionService } from "../../services/orderPromotionService";
import { formatPrice, isValidId, createIdInputHandler, createInputHandler, createSelectHandler } from "../../utils/formatters";

interface PromotionTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

interface CartItem {
  id: number;
  name: string;
  priceInCents: number;
  quantity: number;
}

export const PromotionTab = ({ onMessage, onRefreshLogs }: PromotionTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [promotionName, setPromotionName] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundPromotion, setFoundPromotion] = useState<Promotion | null>(null);
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
      setFoundPromotion(null);
      return;
    }

    try {
      const promotion = await promotionService.getById(parseInt(recordId, 10));
      setFoundPromotion(promotion);
      onMessage(`Found Promotion #${promotion.id}: ${promotion.name} - $${formatPrice(promotion.totalPrice)} (${promotion.itemCount} items)`);
      onRefreshLogs();
    } catch (err) {
      setFoundPromotion(null);
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      onMessage("Error: Please enter a valid record ID");
      return;
    }

    try {
      await promotionService.delete(parseInt(deleteId, 10));
      onMessage(`Successfully deleted promotion with ID ${deleteId}`);
      setDeleteId("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleShowItems = async () => {
    if (!foundPromotion || !foundPromotion.itemIDs || foundPromotion.itemIDs.length === 0) {
      onMessage("No items to display");
      return;
    }

    try {
      const fetchedItems = await Promise.all(
        foundPromotion.itemIDs.map((id) => itemService.getById(id))
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

  const handleCreatePromotion = async () => {
    if (!promotionName || promotionName.trim().length === 0) {
      onMessage("Error: Please enter a promotion name");
      return;
    }

    if (cart.length === 0) {
      onMessage("Error: Please add at least one item to the promotion");
      return;
    }

    try {
      const itemIDs: number[] = [];
      cart.forEach((item) => {
        for (let i = 0; i < item.quantity; i++) {
          itemIDs.push(item.id);
        }
      });

      const promotionId = await orderPromotionService.createPromotion(promotionName, itemIDs);
      onMessage(`Promotion #${promotionId} created successfully: ${promotionName} ($${formatPrice(calculateTotal())})`);
      setPromotionName("");
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
                onChange={createSelectHandler(setSelectedItemId)}
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
                  id="promotion-name"
                  placeholder="Promotion Name"
                  value={promotionName}
                  onChange={createInputHandler(setPromotionName)}
                />
                <Button variant="primary" onClick={handleCreatePromotion}>
                  Create Promotion
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

          {foundPromotion && (
            <div className="details-card">
              <h3>Promotion Details</h3>
              <div className="details-content">
                <div className="details-row">
                  <span className="details-label">ID:</span>
                  <span className="details-value">{foundPromotion.id}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Name:</span>
                  <span className="details-value">{foundPromotion.name}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Total Price:</span>
                  <span className="details-value">${formatPrice(foundPromotion.totalPrice)}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Item Count:</span>
                  <span className="details-value">{foundPromotion.itemCount}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Item IDs:</span>
                  <span className="details-value" onClick={handleShowItems} style={{ cursor: "pointer" }}>
                    {foundPromotion.itemIDs.join(", ")}
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

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} title="Promotion Items">
        <ItemList items={items} />
      </Modal>
    </>
  );
};

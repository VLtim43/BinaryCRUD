import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { DataTable } from "../DataTable";
import { Modal } from "../Modal";
import { orderService, Order } from "../../services/orderService";
import { itemService, Item } from "../../services/itemService";
import { formatPrice, isValidId, createIdInputHandler } from "../../utils/formatters";

interface OrderTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const OrderTab = ({ onMessage, onRefreshLogs }: OrderTabProps) => {
  const [subTab, setSubTab] = useState<"read" | "delete">("read");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundOrder, setFoundOrder] = useState<Order | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [items, setItems] = useState<Item[]>([]);

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

  return (
    <>
      <div className="sub_tabs">
        <Button className={`tab ${subTab === "read" ? "active" : ""}`} onClick={() => setSubTab("read")}>
          Read
        </Button>
        <Button className={`tab ${subTab === "delete" ? "active" : ""}`} onClick={() => setSubTab("delete")}>
          Delete
        </Button>
      </div>

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

import { h } from "preact";
import { useState } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { itemService, Item } from "../../services/itemService";
import { useMessage } from "../../hooks/useMessage";
import { formatPrice, parsePrice, isValidPrice, isValidId, createIdInputHandler } from "../../utils/formatters";

interface ItemTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const ItemTab = ({ onMessage, onRefreshLogs }: ItemTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [itemName, setItemName] = useState("");
  const [itemPrice, setItemPrice] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundItem, setFoundItem] = useState<Item | null>(null);

  const handleCreate = async () => {
    if (!itemName || itemName.trim().length === 0) {
      onMessage("Error: Cannot add empty item");
      return;
    }

    if (!isValidPrice(itemPrice)) {
      onMessage("Error: Please enter a valid price");
      return;
    }

    try {
      const priceInCents = parsePrice(itemPrice);
      await itemService.create(itemName, priceInCents);
      onMessage(`Item saved: ${itemName} ($${itemPrice})`);
      setItemName("");
      setItemPrice("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleRead = async () => {
    if (!isValidId(recordId)) {
      onMessage("Error: Please enter a valid record ID");
      setFoundItem(null);
      return;
    }

    try {
      const item = await itemService.getById(parseInt(recordId, 10));
      setFoundItem(item);
      onMessage(`Found Item #${item.id}: ${item.name} - $${formatPrice(item.priceInCents)}`);
      onRefreshLogs();
    } catch (err) {
      setFoundItem(null);
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      onMessage("Error: Please enter a valid record ID");
      return;
    }

    try {
      await itemService.delete(parseInt(deleteId, 10));
      onMessage(`Successfully deleted item with ID ${deleteId}`);
      setDeleteId("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handlePriceChange = (e: Event) => {
    const target = e.target as HTMLInputElement;
    const value = target.value;
    if (value === "" || /^\d*\.?\d{0,2}$/.test(value)) {
      setItemPrice(value);
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
        <div className="input-box">
          <Input
            id="name"
            placeholder="Item Name"
            value={itemName}
            onChange={(e: Event) => {
              const target = e.target as HTMLInputElement;
              setItemName(target.value);
            }}
          />
          <Input
            id="price"
            placeholder="Price ($)"
            value={itemPrice}
            onChange={handlePriceChange}
          />
          <Button onClick={handleCreate}>Add Item</Button>
        </div>
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

          {foundItem && (
            <div className="details-card">
              <h3>Item Details</h3>
              <div className="details-content">
                <div className="details-row">
                  <span className="details-label">ID:</span>
                  <span className="details-value">{foundItem.id}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Name:</span>
                  <span className="details-value">{foundItem.name}</span>
                </div>
                <div className="details-row">
                  <span className="details-label">Price:</span>
                  <span className="details-value">${formatPrice(foundItem.priceInCents)}</span>
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
    </>
  );
};

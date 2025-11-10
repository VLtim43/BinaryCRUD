import { h } from "preact";
import { useState } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { itemService, Item } from "../../services/itemService";
import { useMessage } from "../../hooks/useMessage";
import { formatPrice, parsePrice, isValidPrice, isValidId } from "../../utils/formatters";

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
    } catch (err: any) {
      onMessage(`Error: ${err}`);
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
    } catch (err: any) {
      setFoundItem(null);
      onMessage(`Error: ${err}`);
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
    } catch (err: any) {
      onMessage(`Error: ${err}`);
    }
  };

  const handlePriceChange = (e: any) => {
    const value = e.target.value;
    if (value === "" || /^\d*\.?\d{0,2}$/.test(value)) {
      setItemPrice(value);
    }
  };

  const handleIdChange = (e: any, setter: (value: string) => void) => {
    const value = e.target.value;
    if (value === "" || /^\d+$/.test(value)) {
      setter(value);
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
            onChange={(e: any) => setItemName(e.target.value)}
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
              onChange={(e: any) => handleIdChange(e, setRecordId)}
            />
            <Button onClick={handleRead}>Get Record</Button>
          </div>

          {foundItem && (
            <div
              style={{
                marginTop: "20px",
                padding: "20px",
                backgroundColor: "rgba(255, 255, 255, 0.1)",
                borderRadius: "8px",
                border: "1px solid rgba(255, 255, 255, 0.2)",
              }}
            >
              <h3 style={{ margin: "0 0 15px 0", color: "#fff" }}>Item Details</h3>
              <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "8px",
                    backgroundColor: "rgba(0, 0, 0, 0.2)",
                    borderRadius: "4px",
                  }}
                >
                  <span style={{ color: "#aaa", fontWeight: "bold" }}>ID:</span>
                  <span style={{ color: "#fff" }}>{foundItem.id}</span>
                </div>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "8px",
                    backgroundColor: "rgba(0, 0, 0, 0.2)",
                    borderRadius: "4px",
                  }}
                >
                  <span style={{ color: "#aaa", fontWeight: "bold" }}>Name:</span>
                  <span style={{ color: "#fff" }}>{foundItem.name}</span>
                </div>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "8px",
                    backgroundColor: "rgba(0, 0, 0, 0.2)",
                    borderRadius: "4px",
                  }}
                >
                  <span style={{ color: "#aaa", fontWeight: "bold" }}>Price:</span>
                  <span style={{ color: "#fff" }}>${formatPrice(foundItem.priceInCents)}</span>
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
            onChange={(e: any) => handleIdChange(e, setDeleteId)}
          />
          <Button variant="danger" onClick={handleDelete}>
            Delete Record
          </Button>
        </div>
      )}
    </>
  );
};

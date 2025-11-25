import { h } from "preact";
import { useState } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { SubTabs } from "../SubTabs";
import { DeleteForm } from "../DeleteForm";
import { itemService, Item } from "../../services/itemService";
import {
  formatPrice,
  parsePrice,
  isValidPrice,
  isValidId,
  createIdInputHandler,
  formatError,
  CRUD_TABS,
} from "../../utils/formatters";
import { toast } from "../../utils/toast";

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
      toast.warning("Cannot add empty item");
      return;
    }

    if (!isValidPrice(itemPrice)) {
      toast.warning("Please enter a valid price");
      return;
    }

    try {
      const priceInCents = parsePrice(itemPrice);
      await itemService.create(itemName, priceInCents);
      toast.success(`Item saved: ${itemName} ($${itemPrice})`);
      setItemName("");
      setItemPrice("");
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
    }
  };

  const handleRead = async () => {
    if (!isValidId(recordId)) {
      toast.warning("Please enter a valid record ID");
      setFoundItem(null);
      return;
    }

    try {
      const item = await itemService.getById(parseInt(recordId, 10));
      setFoundItem(item);
      onRefreshLogs();
    } catch (err) {
      setFoundItem(null);
      toast.error(formatError(err));
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      toast.warning("Please enter a valid record ID");
      return;
    }

    try {
      await itemService.delete(parseInt(deleteId, 10));
      toast.success(`Item ${deleteId} deleted`);
      setDeleteId("");
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
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
      <SubTabs
        tabs={[...CRUD_TABS]}
        activeTab={subTab}
        onTabChange={(tab) => setSubTab(tab as typeof subTab)}
      />

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
                  <span className="details-value">
                    ${formatPrice(foundItem.priceInCents)}
                  </span>
                </div>
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
    </>
  );
};

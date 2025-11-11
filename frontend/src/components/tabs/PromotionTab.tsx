import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { PromotionCreateForm } from "../PromotionCreateForm";
import { promotionService } from "../../services/promotionService";
import { itemService, Item } from "../../services/itemService";
import { useCart } from "../../hooks/useCart";
import { isValidId, formatPrice, createIdInputHandler } from "../../utils/formatters";

interface PromotionTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const PromotionTab = ({ onMessage, onRefreshLogs }: PromotionTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [promotionName, setPromotionName] = useState("");
  const [readId, setReadId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [availableItems, setAvailableItems] = useState<Item[]>([]);
  const [selectedItemId, setSelectedItemId] = useState("");
  const cart = useCart();

  useEffect(() => {
    loadItems();
  }, []);

  const loadItems = async () => {
    try {
      const items = await itemService.getAll();
      setAvailableItems(items);
    } catch (err) {
      onMessage(`Error loading items: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleAddItem = () => {
    if (!selectedItemId) {
      onMessage("Error: Please select an item");
      return;
    }

    const item = availableItems.find((i) => i.id === parseInt(selectedItemId, 10));
    if (!item) {
      onMessage("Error: Item not found");
      return;
    }

    cart.addItem(item);
    onMessage(`Added ${item.name} to promotion`);
  };

  const handleRemoveItem = (itemId: number) => {
    const item = cart.cart.find((c) => c.id === itemId);
    if (item) {
      cart.removeItem(itemId);
      onMessage(`Removed ${item.name} from promotion`);
    }
  };

  const handleSubmit = async () => {
    if (!promotionName || promotionName.trim().length === 0) {
      onMessage("Error: Please enter a promotion name");
      return;
    }

    if (cart.cart.length === 0) {
      onMessage("Error: Promotion cart is empty");
      return;
    }

    try {
      const itemIDs: number[] = [];
      cart.cart.forEach((item) => {
        for (let i = 0; i < item.quantity; i++) {
          itemIDs.push(item.id);
        }
      });

      const promotionId = await promotionService.create(promotionName, itemIDs);
      onMessage(`Promotion #${promotionId} "${promotionName}" created successfully!`);
      cart.clear();
      setPromotionName("");
      setSelectedItemId("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleRead = async () => {
    if (!isValidId(readId)) {
      onMessage("Error: Please enter a valid promotion ID");
      return;
    }

    try {
      const promotion = await promotionService.getById(parseInt(readId, 10));
      onMessage(
        `Promotion #${promotion.id}: "${promotion.name}" - ${promotion.itemCount} items - Total: $${formatPrice(promotion.totalPrice)}`
      );
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      onMessage("Error: Please enter a valid promotion ID");
      return;
    }

    try {
      await promotionService.delete(parseInt(deleteId, 10));
      onMessage(`Successfully deleted promotion #${deleteId}`);
      setDeleteId("");
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
        <PromotionCreateForm
          promotionName={promotionName}
          onPromotionNameChange={(e: Event) => {
            const target = e.target as HTMLInputElement;
            setPromotionName(target.value);
          }}
          cart={cart.cart}
          availableItems={availableItems}
          selectedItemId={selectedItemId}
          onItemSelect={(e: Event) => {
            const target = e.target as HTMLSelectElement;
            setSelectedItemId(target.value);
          }}
          onAddItem={handleAddItem}
          onRemoveItem={handleRemoveItem}
          onSubmit={handleSubmit}
        />
      )}

      {subTab === "read" && (
        <div className="input-box">
          <Input
            placeholder="Enter Promotion ID"
            value={readId}
            onChange={createIdInputHandler(setReadId)}
          />
          <Button onClick={handleRead}>Get Promotion</Button>
        </div>
      )}

      {subTab === "delete" && (
        <div className="input-box">
          <Input
            placeholder="Enter Promotion ID"
            value={deleteId}
            onChange={createIdInputHandler(setDeleteId)}
          />
          <Button variant="danger" onClick={handleDelete}>
            Delete Promotion
          </Button>
        </div>
      )}
    </>
  );
};

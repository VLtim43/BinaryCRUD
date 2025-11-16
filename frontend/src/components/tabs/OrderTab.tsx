import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { OrderCreateForm } from "../OrderCreateForm";
import { orderService } from "../../services/orderService";
import { itemService, Item } from "../../services/itemService";
import { promotionService, Promotion } from "../../services/promotionService";
import { useCart } from "../../hooks/useCart";
import {
  isValidId,
  formatPrice,
  createIdInputHandler,
} from "../../utils/formatters";

interface OrderTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const OrderTab = ({ onMessage, onRefreshLogs }: OrderTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [customerName, setCustomerName] = useState("");
  const [readId, setReadId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [orderDetails, setOrderDetails] = useState<any>(null);
  const [availableItems, setAvailableItems] = useState<Item[]>([]);
  const [availablePromotions, setAvailablePromotions] = useState<Promotion[]>(
    []
  );
  const [selectedItemId, setSelectedItemId] = useState("");
  const [selectedPromotionId, setSelectedPromotionId] = useState("");
  const [selectedPromotions, setSelectedPromotions] = useState<Promotion[]>([]);
  const cart = useCart();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [items, promotions] = await Promise.all([
        itemService.getAll(),
        promotionService.getAll(),
      ]);
      setAvailableItems(items);
      setAvailablePromotions(promotions);
    } catch (err) {
      onMessage(
        `Error loading data: ${
          err instanceof Error ? err.message : String(err)
        }`
      );
    }
  };

  const handleAddItem = () => {
    if (!selectedItemId) {
      onMessage("Error: Please select an item");
      return;
    }

    const item = availableItems.find(
      (i) => i.id === parseInt(selectedItemId, 10)
    );
    if (!item) {
      onMessage("Error: Item not found");
      return;
    }

    cart.addItem(item);
    onMessage(`Added ${item.name} to cart`);
  };

  const handleRemoveItem = (itemId: number) => {
    const item = cart.cart.find((c) => c.id === itemId);
    if (item) {
      cart.removeItem(itemId);
      onMessage(`Removed ${item.name} from cart`);
    }
  };

  const handleAddPromotion = () => {
    if (!selectedPromotionId) {
      onMessage("Error: Please select a promotion");
      return;
    }

    const promo = availablePromotions.find(
      (p) => p.id === parseInt(selectedPromotionId, 10)
    );
    if (!promo) {
      onMessage("Error: Promotion not found");
      return;
    }

    if (selectedPromotions.some((p) => p.id === promo.id)) {
      onMessage("Error: Promotion already added");
      return;
    }

    setSelectedPromotions([...selectedPromotions, promo]);
    setSelectedPromotionId("");
    onMessage(`Added promotion: ${promo.name}`);
  };

  const handleRemovePromotion = (promoId: number) => {
    const promo = selectedPromotions.find((p) => p.id === promoId);
    if (promo) {
      setSelectedPromotions(selectedPromotions.filter((p) => p.id !== promoId));
      onMessage(`Removed promotion: ${promo.name}`);
    }
  };

  const handleSubmit = async () => {
    if (!customerName || customerName.trim().length === 0) {
      onMessage("Error: Please enter a customer name");
      return;
    }

    if (cart.cart.length === 0) {
      onMessage("Error: Cart is empty");
      return;
    }

    try {
      const itemIDs: number[] = [];
      cart.cart.forEach((item) => {
        for (let i = 0; i < item.quantity; i++) {
          itemIDs.push(item.id);
        }
      });

      const orderId = await orderService.create(customerName, itemIDs);

      if (selectedPromotions.length > 0) {
        for (const promo of selectedPromotions) {
          await promotionService.applyToOrder(orderId, promo.id);
        }
        onMessage(
          `Order #${orderId} created for ${customerName} with ${selectedPromotions.length} promotion(s)!`
        );
      } else {
        onMessage(
          `Order #${orderId} created successfully for ${customerName}!`
        );
      }

      cart.clear();
      setCustomerName("");
      setSelectedItemId("");
      setSelectedPromotions([]);
      setSelectedPromotionId("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleRead = async () => {
    if (!isValidId(readId)) {
      onMessage("Error: Please enter a valid order ID");
      setOrderDetails(null);
      return;
    }

    try {
      const order = await orderService.getByIdWithPromotions(
        parseInt(readId, 10)
      );
      setOrderDetails(order);
      onMessage(
        `Order #${order.id}: ${order.customerName} - ${
          order.itemCount
        } items - Total: $${formatPrice(order.totalPrice)}`
      );
      onRefreshLogs();
    } catch (err) {
      setOrderDetails(null);
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      onMessage("Error: Please enter a valid order ID");
      return;
    }

    try {
      await orderService.delete(parseInt(deleteId, 10));
      onMessage(`Successfully deleted order #${deleteId}`);
      setDeleteId("");
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  return (
    <>
      <div className="sub_tabs">
        <Button
          className={`tab ${subTab === "create" ? "active" : ""}`}
          onClick={() => setSubTab("create")}
        >
          Create
        </Button>
        <Button
          className={`tab ${subTab === "read" ? "active" : ""}`}
          onClick={() => setSubTab("read")}
        >
          Read
        </Button>
        <Button
          className={`tab ${subTab === "delete" ? "active" : ""}`}
          onClick={() => setSubTab("delete")}
        >
          Delete
        </Button>
      </div>

      {subTab === "delete" && (
        <div className="input-box">
          <Input
            placeholder="Enter Order ID"
            value={deleteId}
            onChange={createIdInputHandler(setDeleteId)}
          />
          <Button variant="danger" onClick={handleDelete}>
            Delete Order
          </Button>
        </div>
      )}
    </>
  );
};

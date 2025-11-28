import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { Select } from "../Select";
import { Modal } from "../Modal";
import { ItemList } from "../ItemList";
import { SubTabs } from "../SubTabs";
import { DeleteForm } from "../DeleteForm";
import { promotionService, Promotion } from "../../services/promotionService";
import { itemService, Item } from "../../services/itemService";
import { orderPromotionService } from "../../services/orderPromotionService";
import {
  formatPrice,
  isValidId,
  createIdInputHandler,
  createInputHandler,
  createSelectHandler,
  formatError,
  COMPACT_SELECT_STYLE,
  CRUD_TABS,
} from "../../utils/formatters";
import { toast } from "../../utils/toast";
import { useCart } from "../../hooks/useCart";
import { useItems } from "../../hooks/useItems";

interface PromotionTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const PromotionTab = ({
  onMessage,
  onRefreshLogs,
}: PromotionTabProps) => {
  const [subTab, setSubTab] = useState<"create" | "read" | "delete">("create");
  const [promotionName, setPromotionName] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundPromotion, setFoundPromotion] = useState<Promotion | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [items, setItems] = useState<Item[]>([]);

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
    }
  }, [subTab]);

  const handleRead = async () => {
    if (!isValidId(recordId)) {
      toast.warning("Please enter a valid record ID");
      setFoundPromotion(null);
      return;
    }

    try {
      const promotion = await promotionService.getById(parseInt(recordId, 10));
      setFoundPromotion(promotion);
      onRefreshLogs();
    } catch (err) {
      setFoundPromotion(null);
      toast.error(formatError(err));
    }
  };

  const handleDelete = async () => {
    if (!isValidId(deleteId)) {
      toast.warning("Please enter a valid record ID");
      return;
    }

    try {
      await promotionService.delete(parseInt(deleteId, 10));
      toast.success(`Promotion ${deleteId} deleted`);
      setDeleteId("");
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
    }
  };

  const handleShowItems = async () => {
    if (
      !foundPromotion ||
      !foundPromotion.itemIDs ||
      foundPromotion.itemIDs.length === 0
    ) {
      toast.warning("No items to display");
      return;
    }

    try {
      const fetchedItems = await Promise.all(
        foundPromotion.itemIDs.map(async (id) => {
          try {
            return await itemService.getById(id);
          } catch {
            return {
              id,
              name: "[Deleted Item]",
              priceInCents: 0,
              isDeleted: true,
            };
          }
        })
      );
      setItems(fetchedItems);
      setIsModalOpen(true);
      onRefreshLogs();
    } catch (err) {
      toast.error("Failed to fetch items");
    }
  };

  const handleCreatePromotion = async () => {
    if (!promotionName || promotionName.trim().length === 0) {
      toast.warning("Please enter a promotion name");
      return;
    }

    if (cart.length === 0) {
      toast.warning("Please add at least one item to the promotion");
      return;
    }

    try {
      const itemIDs = getItemIDs();
      const promotionId = await orderPromotionService.createPromotion(
        promotionName,
        itemIDs
      );
      toast.success(`Promotion #${promotionId} created: ${promotionName}`);
      setPromotionName("");
      clearCart();
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
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
            <div className="cart-header">
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

            <div className="cart-total">
              Total: ${formatPrice(calculateTotal())} ({getTotalItemCount()}{" "}
              items)
            </div>

            <div className="cart-items">
              {cart.length === 0 ? (
                <div className="cart-empty">No items added yet</div>
              ) : (
                cart.map((item) => (
                  <div key={item.id} className="cart-item">
                    <div className="cart-item-info">
                      <div className="cart-item-name">{item.name}</div>
                      <div className="cart-item-id">
                        ID: {item.id} | ${formatPrice(item.priceInCents)} each
                      </div>
                    </div>
                    <div className="cart-item-controls">
                      <div className="cart-item-quantity">x{item.quantity}</div>
                      <Button
                        size="small"
                        variant="danger"
                        onClick={() => removeFromCart(item.id)}
                      >
                        ×
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </div>

            <div className="cart-footer">
              <div
                className="input-box"
                style={COMPACT_SELECT_STYLE}
              >
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
                  <span className="details-value">
                    ${formatPrice(foundPromotion.totalPrice)}
                  </span>
                </div>
                <div className="details-row" onClick={handleShowItems} style={{ cursor: "pointer" }}>
                  <span className="details-label">Items:</span>
                  <span className="details-value">{foundPromotion.itemCount} items (click to view)</span>
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

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={foundPromotion ? `${foundPromotion.name} - $${formatPrice(foundPromotion.totalPrice)}` : "Promotion Items"}
      >
        <ItemList items={items} />
      </Modal>
    </>
  );
};

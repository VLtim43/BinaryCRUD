import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { DataTable } from "../DataTable";
import { promotionService, Promotion } from "../../services/promotionService";
import { formatPrice, isValidId, createIdInputHandler } from "../../utils/formatters";

interface PromotionTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const PromotionTab = ({ onMessage, onRefreshLogs }: PromotionTabProps) => {
  const [subTab, setSubTab] = useState<"read" | "delete">("read");
  const [recordId, setRecordId] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [foundPromotion, setFoundPromotion] = useState<Promotion | null>(null);

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
                  <span className="details-value">{foundPromotion.itemIDs.join(", ")}</span>
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

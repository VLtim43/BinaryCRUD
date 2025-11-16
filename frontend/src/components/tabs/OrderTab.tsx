import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Input } from "../Input";
import { DataTable } from "../DataTable";
import { orderService, Order } from "../../services/orderService";
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
  const [allOrders, setAllOrders] = useState<Order[]>([]);

  useEffect(() => {
    loadAllOrders();
  }, []);

  const loadAllOrders = async () => {
    try {
      const orders = await orderService.getAll();
      setAllOrders(orders);
    } catch (err) {
      console.error("Error loading orders:", err);
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
      await loadAllOrders();
      onRefreshLogs();
    } catch (err) {
      onMessage(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  return (
    <>
      {allOrders.length > 0 && (
        <div className="details-card max-height-300" style={{ marginBottom: "20px" }}>
          <h3>All Orders ({allOrders.length})</h3>
          <DataTable
            columns={[
              { key: "id", header: "ID", align: "left", minWidth: "60px" },
              { key: "customer", header: "Customer", align: "left", minWidth: "150px" },
              {
                key: "totalPrice",
                header: "Total Price",
                align: "right",
                minWidth: "100px",
                render: (value) => `$${formatPrice(value)}`,
              },
              { key: "itemCount", header: "Items", align: "center", minWidth: "80px" },
            ]}
            data={allOrders}
            maxHeight="200px"
            minWidth="500px"
          />
        </div>
      )}

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
                  <span className="details-value">{foundOrder.itemIDs.join(", ")}</span>
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

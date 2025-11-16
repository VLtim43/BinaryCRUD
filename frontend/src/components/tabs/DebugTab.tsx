import { h } from "preact";
import { useState } from "preact/hooks";
import { Button } from "../Button";
import { DataTable, TableColumn } from "../DataTable";
import { systemService } from "../../services/systemService";
import { itemService, Item } from "../../services/itemService";
import { orderService, Order } from "../../services/orderService";
import { promotionService, Promotion } from "../../services/promotionService";
import { formatPrice } from "../../utils/formatters";

interface DebugTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
  subTab: "tools" | "print";
  onSubTabChange: (subTab: "tools" | "print") => void;
}

export const DebugTab = ({ onMessage, onRefreshLogs, subTab, onSubTabChange }: DebugTabProps) => {
  const [indexData, setIndexData] = useState<any>(null);
  const [printData, setPrintData] = useState<{
    items?: Item[];
    orders?: Order[];
    promotions?: Promotion[];
  }>({});

  const handlePopulateClick = async () => {
    try {
      onMessage("Populating all data...");
      await systemService.populateInventory();
      onMessage("All data populated successfully! Check logs for details.");
      onRefreshLogs();
    } catch (err: any) {
      onMessage(`Error: ${err}`);
      onRefreshLogs();
    }
  };

  const handlePrintIndex = async () => {
    onMessage("Loading index contents...");
    try {
      const data = await systemService.getIndexContents();
      setIndexData(data);
      onMessage(`Index loaded: ${data.count} entries. Scroll down to see details.`);
      onRefreshLogs();
    } catch (err: any) {
      setIndexData(null);
      onMessage(`Error loading index: ${err}`);
    }
  };

  const handleDeleteAll = async () => {
    try {
      await systemService.deleteAllFiles();
      setIndexData(null);
      onMessage("All generated files deleted successfully!");
      onRefreshLogs();
    } catch (err: any) {
      onMessage(`Error: ${err}`);
    }
  };

  const handlePrintAllItems = async () => {
    try {
      onMessage("Loading all items...");
      const items = await itemService.getAll();
      setPrintData({ items });
      onMessage(`Loaded ${items.length} items`);
    } catch (err: any) {
      onMessage(`Error loading items: ${err}`);
    }
  };

  const handlePrintAllOrders = async () => {
    try {
      onMessage("Loading all orders...");
      const orders = await orderService.getAll();
      setPrintData({ orders });
      onMessage(`Loaded ${orders.length} orders`);
    } catch (err: any) {
      onMessage(`Error loading orders: ${err}`);
    }
  };

  const handlePrintAllPromotions = async () => {
    try {
      onMessage("Loading all promotions...");
      const promotions = await promotionService.getAll();
      setPrintData({ promotions });
      onMessage(`Loaded ${promotions.length} promotions`);
    } catch (err: any) {
      onMessage(`Error loading promotions: ${err}`);
    }
  };

  return (
    <>
      <div className="sub_tabs">
        <Button className={`tab ${subTab === "tools" ? "active" : ""}`} onClick={() => onSubTabChange("tools")}>
          Tools
        </Button>
        <Button className={`tab ${subTab === "print" ? "active" : ""}`} onClick={() => onSubTabChange("print")}>
          Print
        </Button>
      </div>

      {subTab === "tools" && (
        <>
          <div className="input-box">
            <Button onClick={handlePopulateClick}>Populate All Data</Button>
            <Button onClick={handlePrintIndex}>Print Index</Button>
            <Button variant="danger" onClick={handleDeleteAll}>
              Delete All Files
            </Button>
          </div>

          {indexData && (
            <div className="details-card max-height-400">
              <h3>B+ Tree Index Contents</h3>
              <div className="details-info">Total entries: {indexData.count}</div>
              <DataTable
                columns={[
                  { key: "id", header: "Item ID", align: "left" },
                  {
                    key: "offset",
                    header: "File Offset",
                    align: "left",
                    render: (value) => <span className="data-table-monospace">{value} bytes</span>,
                  },
                ]}
                data={indexData.entries}
                maxHeight="300px"
              />
            </div>
          )}
        </>
      )}

      {subTab === "print" && (
        <>
          <div className="input-box">
            <Button onClick={handlePrintAllItems}>Print All Items</Button>
            <Button onClick={handlePrintAllOrders}>Print All Orders</Button>
            <Button onClick={handlePrintAllPromotions}>Print All Promotions</Button>
          </div>

          {printData.items && (
            <div className="details-card max-height-300">
              <h3>All Items ({printData.items.length})</h3>
              <DataTable
                columns={[
                  { key: "id", header: "ID", align: "left", minWidth: "60px" },
                  { key: "name", header: "Name", align: "left", minWidth: "200px" },
                  {
                    key: "priceInCents",
                    header: "Price",
                    align: "right",
                    minWidth: "100px",
                    render: (value) => `$${formatPrice(value)}`,
                  },
                ]}
                data={printData.items}
                maxHeight="220px"
                minWidth="400px"
              />
            </div>
          )}

          {printData.orders && (
            <div className="details-card max-height-300">
              <h3>All Orders ({printData.orders.length})</h3>
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
                data={printData.orders}
                maxHeight="220px"
                minWidth="500px"
              />
            </div>
          )}

          {printData.promotions && (
            <div className="details-card max-height-300">
              <h3>All Promotions ({printData.promotions.length})</h3>
              <DataTable
                columns={[
                  { key: "id", header: "ID", align: "left", minWidth: "60px" },
                  { key: "name", header: "Name", align: "left", minWidth: "150px" },
                  {
                    key: "totalPrice",
                    header: "Total Price",
                    align: "right",
                    minWidth: "100px",
                    render: (value) => `$${formatPrice(value)}`,
                  },
                  { key: "itemCount", header: "Items", align: "center", minWidth: "80px" },
                ]}
                data={printData.promotions}
                maxHeight="220px"
                minWidth="500px"
              />
            </div>
          )}
        </>
      )}
    </>
  );
};

import { h } from "preact";
import { useState } from "preact/hooks";
import { Button } from "../Button";
import { Modal } from "../Modal";
import { ItemList } from "../ItemList";
import { DataTable, TableColumn } from "../DataTable";
import { systemService } from "../../services/systemService";
import { itemService, Item } from "../../services/itemService";
import { orderService, Order } from "../../services/orderService";
import { promotionService, Promotion } from "../../services/promotionService";
import { orderPromotionService, OrderWithPromotions } from "../../services/orderPromotionService";
import { formatPrice, PROMO_CARD_STYLE } from "../../utils/formatters";

interface DebugTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
  subTab: "tools" | "print";
  onSubTabChange: (subTab: "tools" | "print") => void;
}

export const DebugTab = ({ onMessage, onRefreshLogs, subTab, onSubTabChange }: DebugTabProps) => {
  const [indexData, setIndexData] = useState<{
    items?: any;
    orders?: any;
    promotions?: any;
  }>({});
  const [printData, setPrintData] = useState<{
    items?: Item[];
    orders?: Order[];
    promotions?: Promotion[];
  }>({});
  const [isItemModalOpen, setIsItemModalOpen] = useState(false);
  const [isPromoModalOpen, setIsPromoModalOpen] = useState(false);
  const [items, setItems] = useState<Item[]>([]);
  const [promoItems, setPromoItems] = useState<Item[]>([]);
  const [selectedOrderForView, setSelectedOrderForView] = useState<OrderWithPromotions | null>(null);
  const [selectedPromoForView, setSelectedPromoForView] = useState<{ id: number; name: string } | null>(null);

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

  const handlePrintItemIndex = async () => {
    onMessage("Loading item index contents...");
    try {
      const data = await systemService.getIndexContents();
      setPrintData({});
      setIndexData({ items: data });
      onMessage(`Item index loaded: ${data.count} entries.`);
      onRefreshLogs();
    } catch (err: any) {
      setIndexData({});
      onMessage(`Error loading item index: ${err}`);
    }
  };

  const handlePrintOrderIndex = async () => {
    onMessage("Loading order index contents...");
    try {
      const data = await systemService.getOrderIndexContents();
      setPrintData({});
      setIndexData({ orders: data });
      onMessage(`Order index loaded: ${data.count} entries.`);
      onRefreshLogs();
    } catch (err: any) {
      setIndexData({});
      onMessage(`Error loading order index: ${err}`);
    }
  };

  const handlePrintPromotionIndex = async () => {
    onMessage("Loading promotion index contents...");
    try {
      const data = await systemService.getPromotionIndexContents();
      setPrintData({});
      setIndexData({ promotions: data });
      onMessage(`Promotion index loaded: ${data.count} entries.`);
      onRefreshLogs();
    } catch (err: any) {
      setIndexData({});
      onMessage(`Error loading promotion index: ${err}`);
    }
  };

  const handleDeleteAll = async () => {
    try {
      await systemService.deleteAllFiles();
      setIndexData({});
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
      setIndexData({});
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
      setIndexData({});
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
      setIndexData({});
      setPrintData({ promotions });
      onMessage(`Loaded ${promotions.length} promotions`);
    } catch (err: any) {
      onMessage(`Error loading promotions: ${err}`);
    }
  };

  const handleShowOrderItems = async (orderId: number) => {
    try {
      const order = await orderPromotionService.getOrderWithPromotions(orderId);
      setSelectedOrderForView(order);

      if (!order.itemIDs || order.itemIDs.length === 0) {
        onMessage("No items in this order");
        return;
      }

      const fetchedItems = await Promise.all(
        order.itemIDs.map((id) => itemService.getById(id))
      );
      setItems(fetchedItems);
      setIsItemModalOpen(true);
      onRefreshLogs();
    } catch (err: any) {
      onMessage(`Error fetching order items: ${err}`);
    }
  };

  const handleShowPromotionItems = async (promotionId: number, promotionName: string) => {
    try {
      const promotion = await promotionService.getById(promotionId);
      if (!promotion.itemIDs || promotion.itemIDs.length === 0) {
        onMessage("No items in this promotion");
        return;
      }

      const fetchedItems = await Promise.all(
        promotion.itemIDs.map((id) => itemService.getById(id))
      );
      setPromoItems(fetchedItems);
      setSelectedPromoForView({ id: promotionId, name: promotionName });
      setIsPromoModalOpen(true);
      onRefreshLogs();
    } catch (err: any) {
      onMessage(`Error fetching promotion items: ${err}`);
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
            <Button variant="danger" onClick={handleDeleteAll}>
              Delete All Files
            </Button>
          </div>
        </>
      )}

      {subTab === "print" && (
        <>
          <div className="button-grid">
            <div className="button-grid-label">Data</div>
            <Button onClick={handlePrintAllItems}>Print All Items</Button>
            <Button onClick={handlePrintAllOrders}>Print All Orders</Button>
            <Button onClick={handlePrintAllPromotions}>Print All Promotions</Button>
            <div className="button-grid-label">Indexes</div>
            <Button onClick={handlePrintItemIndex}>Print Item Index</Button>
            <Button onClick={handlePrintOrderIndex}>Print Order Index</Button>
            <Button onClick={handlePrintPromotionIndex}>Print Promotion Index</Button>
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
                  {
                    key: "id",
                    header: "ID",
                    align: "left",
                    minWidth: "60px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowOrderItems(row.id)}
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    )
                  },
                  {
                    key: "customer",
                    header: "Customer",
                    align: "left",
                    minWidth: "150px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowOrderItems(row.id)}
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    )
                  },
                  {
                    key: "totalPrice",
                    header: "Total Price",
                    align: "right",
                    minWidth: "100px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowOrderItems(row.id)}
                        style={{ cursor: "pointer" }}
                      >
                        ${formatPrice(value)}
                      </span>
                    )
                  },
                  {
                    key: "itemCount",
                    header: "Items",
                    align: "center",
                    minWidth: "80px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowOrderItems(row.id)}
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    )
                  },
                ]}
                data={printData.orders}
                maxHeight="220px"
                minWidth="400px"
              />
            </div>
          )}

          {printData.promotions && (
            <div className="details-card max-height-300">
              <h3>All Promotions ({printData.promotions.length})</h3>
              <DataTable
                columns={[
                  {
                    key: "id",
                    header: "ID",
                    align: "left",
                    minWidth: "60px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowPromotionItems(row.id, row.name)}
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    )
                  },
                  {
                    key: "name",
                    header: "Name",
                    align: "left",
                    minWidth: "150px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowPromotionItems(row.id, row.name)}
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    )
                  },
                  {
                    key: "totalPrice",
                    header: "Total Price",
                    align: "right",
                    minWidth: "100px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowPromotionItems(row.id, row.name)}
                        style={{ cursor: "pointer" }}
                      >
                        ${formatPrice(value)}
                      </span>
                    )
                  },
                  {
                    key: "itemCount",
                    header: "Items",
                    align: "center",
                    minWidth: "80px",
                    render: (value, row) => (
                      <span
                        onClick={() => handleShowPromotionItems(row.id, row.name)}
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    )
                  },
                ]}
                data={printData.promotions}
                maxHeight="220px"
                minWidth="400px"
              />
            </div>
          )}

          {indexData.items && (
            <div className="details-card max-height-300">
              <h3>Item Index ({indexData.items.count} entries)</h3>
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
                data={indexData.items.entries}
                maxHeight="220px"
              />
            </div>
          )}

          {indexData.orders && (
            <div className="details-card max-height-300">
              <h3>Order Index ({indexData.orders.count} entries)</h3>
              <DataTable
                columns={[
                  { key: "id", header: "Order ID", align: "left" },
                  {
                    key: "offset",
                    header: "File Offset",
                    align: "left",
                    render: (value) => <span className="data-table-monospace">{value} bytes</span>,
                  },
                ]}
                data={indexData.orders.entries}
                maxHeight="220px"
              />
            </div>
          )}

          {indexData.promotions && (
            <div className="details-card max-height-300">
              <h3>Promotion Index ({indexData.promotions.count} entries)</h3>
              <DataTable
                columns={[
                  { key: "id", header: "Promotion ID", align: "left" },
                  {
                    key: "offset",
                    header: "File Offset",
                    align: "left",
                    render: (value) => <span className="data-table-monospace">{value} bytes</span>,
                  },
                ]}
                data={indexData.promotions.entries}
                maxHeight="220px"
              />
            </div>
          )}
        </>
      )}

      <Modal isOpen={isItemModalOpen} onClose={() => setIsItemModalOpen(false)} title={selectedOrderForView ? `Order #${selectedOrderForView.id} Items` : "Order Items"}>
        <ItemList items={items}>
          {selectedOrderForView && selectedOrderForView.promotions && selectedOrderForView.promotions.length > 0 && (
            <>
              {selectedOrderForView.promotions.map((promo) => (
                <div
                  key={promo.id}
                  className="cart-item"
                  style={{ ...PROMO_CARD_STYLE, cursor: "pointer" }}
                  onClick={() => {
                    setIsItemModalOpen(false);
                    handleShowPromotionItems(promo.id, promo.name);
                  }}
                >
                  <div className="cart-item-info">
                    <div className="cart-item-name">[PROMO] {promo.name}</div>
                    <div className="cart-item-id">ID: {promo.id} | ${formatPrice(promo.totalPrice)} | {promo.itemCount} items</div>
                  </div>
                </div>
              ))}
            </>
          )}
        </ItemList>
      </Modal>

      <Modal
        isOpen={isPromoModalOpen}
        onClose={() => setIsPromoModalOpen(false)}
        title={selectedPromoForView ? `Promotion: ${selectedPromoForView.name}` : "Promotion Items"}
      >
        <ItemList items={promoItems} />
      </Modal>
    </>
  );
};

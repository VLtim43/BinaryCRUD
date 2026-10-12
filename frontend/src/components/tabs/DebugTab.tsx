import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Button } from "../Button";
import { Modal } from "../Modal";
import { ItemList } from "../ItemList";
import { DataTable } from "../DataTable";
import { SubTabs } from "../SubTabs";
import { systemService } from "../../services/systemService";
import { itemService, Item } from "../../services/itemService";
import { orderService, Order } from "../../services/orderService";
import { promotionService, Promotion } from "../../services/promotionService";
import {
  orderPromotionService,
  OrderWithPromotions,
} from "../../services/orderPromotionService";
import {
  formatPrice,
  formatError,
  createSelectHandler,
  PROMO_CARD_STYLE,
  DEBUG_TABS,
} from "../../utils/formatters";
import { toast } from "../../utils/toast";

type DebugSubTab = "tools" | "print";

interface DebugTabProps {
  onRefreshLogs: () => void;
  onMessage: (msg: string) => void;
  subTab: DebugSubTab;
  onSubTabChange: (subTab: DebugSubTab) => void;
}

type IndexType = "items" | "orders" | "promotions";
type PrintDataType = "items" | "orders" | "promotions";

const DEFAULT_MESSAGE = "Debug Tools and Utilities";

export const DebugTab = ({
  onRefreshLogs,
  onMessage,
  subTab,
  onSubTabChange,
}: DebugTabProps) => {
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
  const [selectedOrderForView, setSelectedOrderForView] =
    useState<OrderWithPromotions | null>(null);
  const [selectedPromoForView, setSelectedPromoForView] = useState<{
    id: number;
    name: string;
    totalPrice: number;
  } | null>(null);
  const [orderItemsPrice, setOrderItemsPrice] = useState<number>(0);

  // Compact state
  const [isCompacting, setIsCompacting] = useState(false);

  const handlePopulateClick = async () => {
    try {
      await systemService.populateInventory();
      toast.success("Inventory populated successfully!");
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
      onRefreshLogs();
    }
  };

  const handleDeleteAll = async () => {
    try {
      await systemService.deleteAllFiles();
      setIndexData({});
      // Individual toasts are emitted from the backend for each folder
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
    }
  };

  const handleCompact = async () => {
    setIsCompacting(true);
    try {
      await systemService.compact();
      // Toast notifications are emitted from the backend
      onRefreshLogs();
    } catch (err) {
      toast.error(formatError(err));
    } finally {
      setIsCompacting(false);
    }
  };

  // Consolidated index loading function
  const handlePrintIndex = async (type: IndexType) => {
    const indexNames: Record<IndexType, string> = {
      items: "item",
      orders: "order",
      promotions: "promotion",
    };
    const indexName = indexNames[type];

    try {
      let data;
      switch (type) {
        case "items":
          data = await systemService.getIndexContents();
          break;
        case "orders":
          data = await systemService.getOrderIndexContents();
          break;
        case "promotions":
          data = await systemService.getPromotionIndexContents();
          break;
      }
      setPrintData({});
      setIndexData({ [type]: data });
      toast.success(`${indexName.charAt(0).toUpperCase() + indexName.slice(1)} index: ${data.count} entries`);
      onRefreshLogs();
    } catch (err) {
      setIndexData({});
      toast.error(`Failed to load ${indexName} index`);
    }
  };

  // Consolidated print all function
  const handlePrintAll = async (type: PrintDataType) => {
    const typeNames: Record<PrintDataType, string> = {
      items: "items",
      orders: "orders",
      promotions: "promotions",
    };
    const typeName = typeNames[type];

    try {
      let data;
      switch (type) {
        case "items":
          data = await itemService.getAll();
          break;
        case "orders":
          data = await orderService.getAll();
          break;
        case "promotions":
          data = await promotionService.getAll();
          break;
      }
      setIndexData({});
      setPrintData({ [type]: data });
      toast.success(`Loaded ${data.length} ${typeName}`);
    } catch (err) {
      toast.error(`Failed to load ${typeName}`);
    }
  };

  const handleShowOrderItems = async (orderId: number) => {
    try {
      const order = await orderPromotionService.getOrderWithPromotions(orderId);
      setSelectedOrderForView(order);

      if (!order.itemIDs || order.itemIDs.length === 0) {
        toast.warning("No items in this order");
        return;
      }

      const fetchedItems = await Promise.all(
        order.itemIDs.map(async (id) => {
          try {
            return await itemService.getById(id);
          } catch {
            // Item was deleted, return placeholder
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
      // Calculate order items price (total minus promotions)
      const promoTotal = order.promotions?.reduce((sum, p) => sum + p.totalPrice, 0) || 0;
      setOrderItemsPrice(order.totalPrice - promoTotal);
      setIsItemModalOpen(true);
      onRefreshLogs();
    } catch (err) {
      toast.error("Failed to fetch order items");
    }
  };

  const handleShowPromotionItems = async (
    promotionId: number,
    promotionName: string,
    totalPrice: number
  ) => {
    try {
      const promotion = await promotionService.getById(promotionId);
      if (!promotion.itemIDs || promotion.itemIDs.length === 0) {
        toast.warning("No items in this promotion");
        return;
      }

      const fetchedItems = await Promise.all(
        promotion.itemIDs.map(async (id) => {
          try {
            return await itemService.getById(id);
          } catch {
            // Item was deleted, return placeholder
            return {
              id,
              name: "[Deleted Item]",
              priceInCents: 0,
              isDeleted: true,
            };
          }
        })
      );
      setPromoItems(fetchedItems);
      setSelectedPromoForView({ id: promotionId, name: promotionName, totalPrice });
      setIsPromoModalOpen(true);
      onRefreshLogs();
    } catch (err) {
      toast.error("Failed to fetch promotion items");
    }
  };

  return (
    <>
      <SubTabs
        tabs={[...DEBUG_TABS]}
        activeTab={subTab}
        onTabChange={(tab) => onSubTabChange(tab as DebugSubTab)}
      />

      {subTab === "tools" && (
        <div className="input-box">
          <Button
            onClick={handlePopulateClick}
            onMouseEnter={() => onMessage("Load sample items, promotions, and orders from seed files")}
            onMouseLeave={() => onMessage(DEFAULT_MESSAGE)}
          >
            Populate Inventory
          </Button>
          <Button
            onClick={handleCompact}
            disabled={isCompacting}
            onMouseEnter={() => onMessage("Permanently remove deleted records and clean orphan references")}
            onMouseLeave={() => onMessage(DEFAULT_MESSAGE)}
          >
            {isCompacting ? "Compacting..." : "Compact Database"}
          </Button>
          <Button
            variant="danger"
            onClick={handleDeleteAll}
            onMouseEnter={() => onMessage("Delete all .bin and .idx files")}
            onMouseLeave={() => onMessage(DEFAULT_MESSAGE)}
          >
            Delete All Files
          </Button>
        </div>
      )}

      {subTab === "print" && (
        <>
          <div className="button-grid">
            <div className="button-grid-label">Data</div>
            <Button onClick={() => handlePrintAll("items")}>
              Print All Items
            </Button>
            <Button onClick={() => handlePrintAll("orders")}>
              Print All Orders
            </Button>
            <Button onClick={() => handlePrintAll("promotions")}>
              Print All Promotions
            </Button>
            <div className="button-grid-label">Indexes</div>
            <Button onClick={() => handlePrintIndex("items")}>
              Print Item Index
            </Button>
            <Button onClick={() => handlePrintIndex("orders")}>
              Print Order Index
            </Button>
            <Button onClick={() => handlePrintIndex("promotions")}>
              Print Promotion Index
            </Button>
          </div>

          {printData.items && (
            <div className="details-card max-height-300">
              <h3>All Items ({printData.items.length})</h3>
              <DataTable
                columns={[
                  { key: "id", header: "ID", align: "left", minWidth: "60px" },
                  {
                    key: "name",
                    header: "Name",
                    align: "left",
                    minWidth: "200px",
                  },
                  {
                    key: "priceInCents",
                    header: "Price",
                    align: "right",
                    minWidth: "100px",
                    render: (value) => `$${formatPrice(value)}`,
                  },
                  {
                    key: "isDeleted",
                    header: "Deleted",
                    align: "center",
                    minWidth: "60px",
                    render: (value) => (value ? "1" : "0"),
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
                    ),
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
                    ),
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
                    ),
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
                    ),
                  },
                  {
                    key: "isDeleted",
                    header: "Deleted",
                    align: "center",
                    minWidth: "60px",
                    render: (value) => (value ? "1" : "0"),
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
                        onClick={() =>
                          handleShowPromotionItems(row.id, row.name, row.totalPrice)
                        }
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    ),
                  },
                  {
                    key: "name",
                    header: "Name",
                    align: "left",
                    minWidth: "150px",
                    render: (value, row) => (
                      <span
                        onClick={() =>
                          handleShowPromotionItems(row.id, row.name, row.totalPrice)
                        }
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    ),
                  },
                  {
                    key: "totalPrice",
                    header: "Total Price",
                    align: "right",
                    minWidth: "100px",
                    render: (value, row) => (
                      <span
                        onClick={() =>
                          handleShowPromotionItems(row.id, row.name, row.totalPrice)
                        }
                        style={{ cursor: "pointer" }}
                      >
                        ${formatPrice(value)}
                      </span>
                    ),
                  },
                  {
                    key: "itemCount",
                    header: "Items",
                    align: "center",
                    minWidth: "80px",
                    render: (value, row) => (
                      <span
                        onClick={() =>
                          handleShowPromotionItems(row.id, row.name, row.totalPrice)
                        }
                        style={{ cursor: "pointer" }}
                      >
                        {value}
                      </span>
                    ),
                  },
                  {
                    key: "isDeleted",
                    header: "Deleted",
                    align: "center",
                    minWidth: "60px",
                    render: (value) => (value ? "1" : "0"),
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
                    render: (value) => (
                      <span className="data-table-monospace">
                        {value} bytes
                      </span>
                    ),
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
                    render: (value) => (
                      <span className="data-table-monospace">
                        {value} bytes
                      </span>
                    ),
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
                    render: (value) => (
                      <span className="data-table-monospace">
                        {value} bytes
                      </span>
                    ),
                  },
                ]}
                data={indexData.promotions.entries}
                maxHeight="220px"
              />
            </div>
          )}
        </>
      )}

      <Modal
        isOpen={isItemModalOpen}
        onClose={() => setIsItemModalOpen(false)}
        title={
          selectedOrderForView
            ? `Order #${selectedOrderForView.id} Items - $${formatPrice(orderItemsPrice)}`
            : "Order Items"
        }
      >
        <ItemList items={items}>
          {selectedOrderForView &&
            selectedOrderForView.promotions &&
            selectedOrderForView.promotions.length > 0 && (
              <>
                {selectedOrderForView.promotions.map((promo) => (
                  <div
                    key={promo.id}
                    className="cart-item"
                    style={{ ...PROMO_CARD_STYLE, cursor: "pointer" }}
                    onClick={() => {
                      setIsItemModalOpen(false);
                      handleShowPromotionItems(promo.id, promo.name, promo.totalPrice);
                    }}
                  >
                    <div className="cart-item-info">
                      <div className="cart-item-name">[PROMO] {promo.name}</div>
                      <div className="cart-item-id">
                        ID: {promo.id} | ${formatPrice(promo.totalPrice)} |{" "}
                        {promo.itemCount} items
                      </div>
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
        title={
          selectedPromoForView
            ? `${selectedPromoForView.name} - $${formatPrice(selectedPromoForView.totalPrice)}`
            : "Promotion Items"
        }
      >
        <ItemList items={promoItems} />
      </Modal>
    </>
  );
};

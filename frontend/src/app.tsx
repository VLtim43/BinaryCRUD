import "./App.scss";
import logo from "./assets/images/logo-universal.png";
import {
  AddItem,
  GetItem,
  DeleteItem,
  DeleteAllFiles,
  GetLogs,
  ClearLogs,
  PopulateInventory,
  GetIndexContents,
  GetAllItems,
  CreateOrder,
  GetOrder,
  DeleteOrder,
  CreatePromotion,
  GetPromotion,
  DeletePromotion,
} from "../wailsjs/go/main/App";
import { Quit } from "../wailsjs/runtime/runtime";
import { useState, useEffect, useRef } from "preact/hooks";
import { h, Fragment } from "preact";

export const App = () => {
  const [activeTab, setActiveTab] = useState<
    "item" | "order" | "promotion" | "debug"
  >("item");
  const [itemSubTab, setItemSubTab] = useState<"create" | "read" | "delete">(
    "create"
  );
  const [orderSubTab, setOrderSubTab] = useState<"create" | "read" | "delete">(
    "create"
  );
  const [promotionSubTab, setPromotionSubTab] = useState<"create" | "read" | "delete">(
    "create"
  );
  const [resultText, setResultText] = useState("Enter item text below 👇");
  const [itemText, setItemText] = useState("");
  const [itemPrice, setItemPrice] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteRecordId, setDeleteRecordId] = useState("");
  const [foundItem, setFoundItem] = useState<{
    id: number;
    name: string;
    priceInCents: number;
  } | null>(null);
  const [availableItems, setAvailableItems] = useState<
    Array<{ id: number; name: string; priceInCents: number }>
  >([]);
  const [cart, setCart] = useState<
    Array<{ id: number; name: string; quantity: number; priceInCents: number }>
  >([]);
  const [selectedItemId, setSelectedItemId] = useState<string>("");
  const [customerName, setCustomerName] = useState("");
  const [orderReadId, setOrderReadId] = useState("");
  const [orderDeleteId, setOrderDeleteId] = useState("");
  const [promotionName, setPromotionName] = useState("");
  const [promotionCart, setPromotionCart] = useState<
    Array<{ id: number; name: string; quantity: number; priceInCents: number }>
  >([]);
  const [selectedPromotionItemId, setSelectedPromotionItemId] = useState<string>("");
  const [promotionReadId, setPromotionReadId] = useState("");
  const [promotionDeleteId, setPromotionDeleteId] = useState("");
  const [useIndex, setUseIndex] = useState(true);
  const [logs, setLogs] = useState<
    Array<{ timestamp: string; level: string; message: string }>
  >([]);
  const [logsPanelOpen, setLogsPanelOpen] = useState(false);
  const [indexData, setIndexData] = useState<any>(null);
  const logsContainerRef = useRef<HTMLDivElement>(null);
  const updateItemText = (e: any) => setItemText(e.target.value);
  const updateItemPrice = (e: any) => {
    const value = e.target.value;
    // Allow empty string, digits, and one decimal point
    if (value === "" || /^\d*\.?\d{0,2}$/.test(value)) {
      setItemPrice(value);
    }
  };
  const updateRecordId = (e: any) => {
    const value = e.target.value;
    // Only allow empty string or non-negative integers
    if (value === "" || /^\d+$/.test(value)) {
      setRecordId(value);
    }
  };
  const updateDeleteRecordId = (e: any) => {
    const value = e.target.value;
    // Only allow empty string or non-negative integers
    if (value === "" || /^\d+$/.test(value)) {
      setDeleteRecordId(value);
    }
  };
  const updateOrderReadId = (e: any) => {
    const value = e.target.value;
    if (value === "" || /^\d+$/.test(value)) {
      setOrderReadId(value);
    }
  };
  const updateOrderDeleteId = (e: any) => {
    const value = e.target.value;
    if (value === "" || /^\d+$/.test(value)) {
      setOrderDeleteId(value);
    }
  };
  const updateCustomerName = (e: any) => setCustomerName(e.target.value);
  const updatePromotionName = (e: any) => setPromotionName(e.target.value);
  const updatePromotionReadId = (e: any) => {
    const value = e.target.value;
    if (value === "" || /^\d+$/.test(value)) {
      setPromotionReadId(value);
    }
  };
  const updatePromotionDeleteId = (e: any) => {
    const value = e.target.value;
    if (value === "" || /^\d+$/.test(value)) {
      setPromotionDeleteId(value);
    }
  };
  const updateResultText = (result: string) => setResultText(result);

  // Get default text for each tab/subtab combination
  const getDefaultText = (
    tab: "item" | "order" | "promotion" | "debug",
    subTab?: "create" | "read" | "delete"
  ) => {
    if (tab === "debug") {
      return "Debug tools and utilities";
    }

    if (tab === "promotion") {
      switch (subTab) {
        case "create":
          return "Create a new promotion by selecting items";
        case "read":
          return "Enter a promotion ID to fetch 👇";
        case "delete":
          return "Enter a promotion ID to delete 👇";
        default:
          return "Create a new promotion by selecting items";
      }
    }

    if (tab === "order") {
      switch (subTab) {
        case "create":
          return "Select items to add to your order";
        case "read":
          return "Enter an order ID to fetch 👇";
        case "delete":
          return "Enter an order ID to delete 👇";
        default:
          return "Select items to add to your order";
      }
    }

    if (tab === "item") {
      switch (subTab) {
        case "create":
          return "Enter item text below 👇";
        case "read":
          return "Enter a record ID to fetch 👇";
        case "delete":
          return "Enter a record ID to delete 👇";
        default:
          return "";
      }
    }

    return "";
  };

  // Handle main tab changes
  const handleTabChange = (tab: "item" | "order" | "promotion" | "debug") => {
    setActiveTab(tab);
    if (tab === "item") {
      // Reset to create subtab
      setItemSubTab("create");
      setResultText(getDefaultText(tab, "create"));
    } else if (tab === "order") {
      // Reset order page state and subtab
      setOrderSubTab("create");
      setSelectedItemId("");
      setCart([]);
      setCustomerName("");
      setResultText(getDefaultText(tab, "create"));
      // Load items when entering order tab
      loadItems();
    } else if (tab === "promotion") {
      // Reset promotion page state and subtab
      setPromotionSubTab("create");
      setSelectedPromotionItemId("");
      setPromotionCart([]);
      setPromotionName("");
      setResultText(getDefaultText(tab, "create"));
      // Load items when entering promotion tab
      loadItems();
    } else {
      setResultText(getDefaultText(tab));
    }
  };

  // Handle item subtab changes
  const handleItemSubTabChange = (subTab: "create" | "read" | "delete") => {
    setItemSubTab(subTab);
    setResultText(getDefaultText("item", subTab));
    // Clear found item when switching tabs
    setFoundItem(null);
  };

  // Handle order subtab changes
  const handleOrderSubTabChange = (subTab: "create" | "read" | "delete") => {
    setOrderSubTab(subTab);
    setResultText(getDefaultText("order", subTab));
    // Load items when switching to create subtab
    if (subTab === "create") {
      loadItems();
    }
  };

  // Handle promotion subtab changes
  const handlePromotionSubTabChange = (subTab: "create" | "read" | "delete") => {
    setPromotionSubTab(subTab);
    setResultText(getDefaultText("promotion", subTab));
    // Load items when switching to create subtab
    if (subTab === "create") {
      loadItems();
    }
  };

  const addItem = () => {
    // Validate input before sending to backend
    if (!itemText || itemText.trim().length === 0) {
      updateResultText("Error: Cannot add empty item");
      return;
    }

    if (!itemPrice || itemPrice.trim().length === 0) {
      updateResultText("Error: Please enter a price");
      return;
    }

    // Convert price from dollars to cents
    const priceInCents = Math.round(parseFloat(itemPrice) * 100);
    if (isNaN(priceInCents) || priceInCents < 0) {
      updateResultText("Error: Invalid price");
      return;
    }

    AddItem(itemText, priceInCents)
      .then(() => {
        updateResultText(`Item saved: ${itemText} ($${itemPrice})`);
        setItemText("");
        setItemPrice("");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const getRecordById = () => {
    // Validate input
    if (!recordId || recordId.trim().length === 0) {
      updateResultText("Error: Please enter a record ID");
      setFoundItem(null);
      return;
    }

    const id = parseInt(recordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Invalid record ID");
      setFoundItem(null);
      return;
    }

    // Call backend to get item with index preference
    GetItem(id, useIndex)
      .then((item: any) => {
        setFoundItem({
          id: item.id,
          name: item.name,
          priceInCents: item.priceInCents,
        });
        const method = useIndex ? "B+ Tree Index" : "Sequential Scan";
        updateResultText(
          `Found Item #${item.id}: ${item.name} - $${(item.priceInCents / 100).toFixed(2)} (via ${method})`
        );
        refreshLogs();
      })
      .catch((err: any) => {
        setFoundItem(null);
        updateResultText(`Error: ${err}`);
      });
  };

  const deleteAllFiles = () => {
    DeleteAllFiles()
      .then(() => {
        // Clear all cached state
        setAvailableItems([]);
        setCart([]);
        setSelectedItemId("");
        updateResultText("All generated files deleted successfully!");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const populateInventory = () => {
    updateResultText("Populating inventory from items.json...");

    PopulateInventory()
      .then(() => {
        updateResultText("Inventory populated successfully! Check logs for details.");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error populating inventory: ${err}`);
        refreshLogs();
      });
  };

  const printIndex = () => {
    updateResultText("Loading index contents...");

    GetIndexContents()
      .then((data: any) => {
        setIndexData(data);
        updateResultText(
          `Index loaded: ${data.count} entries. Scroll down to see details.`
        );
        refreshLogs();
      })
      .catch((err: any) => {
        setIndexData(null);
        updateResultText(`Error loading index: ${err}`);
      });
  };

  const deleteItem = () => {
    // Validate input
    if (!deleteRecordId || deleteRecordId.trim().length === 0) {
      updateResultText("Error: Please enter a record ID");
      return;
    }

    const id = parseInt(deleteRecordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Invalid record ID");
      return;
    }

    // Call backend to delete item
    DeleteItem(id)
      .then(() => {
        updateResultText(`Successfully deleted item with ID ${id}`);
        setDeleteRecordId("");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Load all items for order/promotion tabs
  const loadItems = () => {
    GetAllItems()
      .then((items: any[]) => {
        setAvailableItems(items);
        updateResultText(`Loaded ${items.length} items`);
      })
      .catch((err: any) => {
        updateResultText(`Error loading items: ${err}`);
        setAvailableItems([]);
      });
  };

  // Add item to cart
  const addToCart = () => {
    if (!selectedItemId) {
      updateResultText("Error: Please select an item");
      return;
    }

    const itemId = parseInt(selectedItemId, 10);
    const item = availableItems.find((i) => i.id === itemId);

    if (!item) {
      updateResultText("Error: Item not found");
      return;
    }

    // Check if item already in cart
    const existingItem = cart.find((c) => c.id === itemId);
    if (existingItem) {
      // Increment quantity
      setCart(
        cart.map((c) =>
          c.id === itemId ? { ...c, quantity: c.quantity + 1 } : c
        )
      );
    } else {
      // Add new item to cart
      setCart([
        ...cart,
        {
          id: item.id,
          name: item.name,
          quantity: 1,
          priceInCents: item.priceInCents,
        },
      ]);
    }

    updateResultText(`Added ${item.name} to cart`);
  };

  // Remove item from cart
  const removeFromCart = (itemId: number) => {
    const item = cart.find((c) => c.id === itemId);
    if (item) {
      setCart(cart.filter((c) => c.id !== itemId));
      updateResultText(`Removed ${item.name} from cart`);
    }
  };

  // Submit order - writes order to orders.bin
  const submitOrder = () => {
    if (!customerName || customerName.trim().length === 0) {
      updateResultText("Error: Please enter a customer name");
      return;
    }

    if (cart.length === 0) {
      updateResultText("Error: Cart is empty");
      return;
    }

    // Convert cart to item IDs array (with duplicates for quantity)
    const itemIDs: number[] = [];
    cart.forEach(item => {
      for (let i = 0; i < item.quantity; i++) {
        itemIDs.push(item.id);
      }
    });

    CreateOrder(customerName, itemIDs)
      .then((orderId: number) => {
        updateResultText(`Order #${orderId} created successfully for ${customerName}!`);
        // Clear cart and customer name
        setCart([]);
        setCustomerName("");
        setSelectedItemId("");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error creating order: ${err}`);
      });
  };

  // Get order by ID
  const getOrderById = () => {
    if (!orderReadId || orderReadId.trim().length === 0) {
      updateResultText("Error: Please enter an order ID");
      return;
    }

    const id = parseInt(orderReadId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Invalid order ID");
      return;
    }

    GetOrder(id)
      .then((order: any) => {
        updateResultText(
          `Order #${order.id}: ${order.customer} - ${order.itemCount} items - Total: $${(order.totalPrice / 100).toFixed(2)}`
        );
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Delete order by ID
  const deleteOrderById = () => {
    if (!orderDeleteId || orderDeleteId.trim().length === 0) {
      updateResultText("Error: Please enter an order ID");
      return;
    }

    const id = parseInt(orderDeleteId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Invalid order ID");
      return;
    }

    DeleteOrder(id)
      .then(() => {
        updateResultText(`Successfully deleted order #${id}`);
        setOrderDeleteId("");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Add item to promotion cart
  const addToPromotionCart = () => {
    if (!selectedPromotionItemId) {
      updateResultText("Error: Please select an item");
      return;
    }

    const itemId = parseInt(selectedPromotionItemId, 10);
    const item = availableItems.find((i) => i.id === itemId);

    if (!item) {
      updateResultText("Error: Item not found");
      return;
    }

    // Check if item already in cart
    const existingItem = promotionCart.find((c) => c.id === itemId);
    if (existingItem) {
      // Increment quantity
      setPromotionCart(
        promotionCart.map((c) =>
          c.id === itemId ? { ...c, quantity: c.quantity + 1 } : c
        )
      );
    } else {
      // Add new item to cart
      setPromotionCart([
        ...promotionCart,
        {
          id: item.id,
          name: item.name,
          quantity: 1,
          priceInCents: item.priceInCents,
        },
      ]);
    }

    updateResultText(`Added ${item.name} to promotion`);
  };

  // Remove item from promotion cart
  const removeFromPromotionCart = (itemId: number) => {
    const item = promotionCart.find((c) => c.id === itemId);
    if (item) {
      setPromotionCart(promotionCart.filter((c) => c.id !== itemId));
      updateResultText(`Removed ${item.name} from promotion`);
    }
  };

  // Submit promotion
  const submitPromotion = () => {
    if (!promotionName || promotionName.trim().length === 0) {
      updateResultText("Error: Please enter a promotion name");
      return;
    }

    if (promotionCart.length === 0) {
      updateResultText("Error: Promotion cart is empty");
      return;
    }

    // Convert cart to item IDs array (with duplicates for quantity)
    const itemIDs: number[] = [];
    promotionCart.forEach(item => {
      for (let i = 0; i < item.quantity; i++) {
        itemIDs.push(item.id);
      }
    });

    CreatePromotion(promotionName, itemIDs)
      .then((promotionId: number) => {
        updateResultText(`Promotion #${promotionId} "${promotionName}" created successfully!`);
        // Clear cart and promotion name
        setPromotionCart([]);
        setPromotionName("");
        setSelectedPromotionItemId("");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error creating promotion: ${err}`);
      });
  };

  // Get promotion by ID
  const getPromotionById = () => {
    if (!promotionReadId || promotionReadId.trim().length === 0) {
      updateResultText("Error: Please enter a promotion ID");
      return;
    }

    const id = parseInt(promotionReadId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Invalid promotion ID");
      return;
    }

    GetPromotion(id)
      .then((promotion: any) => {
        updateResultText(
          `Promotion #${promotion.id}: "${promotion.name}" - ${promotion.itemCount} items - Total: $${(promotion.totalPrice / 100).toFixed(2)}`
        );
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Delete promotion by ID
  const deletePromotionById = () => {
    if (!promotionDeleteId || promotionDeleteId.trim().length === 0) {
      updateResultText("Error: Please enter a promotion ID");
      return;
    }

    const id = parseInt(promotionDeleteId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Invalid promotion ID");
      return;
    }

    DeletePromotion(id)
      .then(() => {
        updateResultText(`Successfully deleted promotion #${id}`);
        setPromotionDeleteId("");
        refreshLogs();
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Load logs once when panel is opened
  useEffect(() => {
    if (logsPanelOpen) {
      refreshLogs();
    }
  }, [logsPanelOpen]);

  // Auto-scroll to bottom when logs update
  useEffect(() => {
    if (logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [logs]);

  // Refresh logs from backend
  const refreshLogs = () => {
    GetLogs()
      .then((newLogs: Array<{ timestamp: string; level: string; message: string }>) => {
        console.log("Logs fetched:", newLogs);
        setLogs(newLogs);
      })
      .catch((err: any) => {
        console.error("Error loading logs:", err);
      });
  };

  // Clear all logs
  const clearLogs = () => {
    ClearLogs()
      .then(() => {
        setLogs([]);
        updateResultText("Logs cleared");
      })
      .catch((err: any) => {
        updateResultText(`Error clearing logs: ${err}`);
      });
  };

  // Copy logs to clipboard
  const copyLogs = () => {
    const logText = logs
      .map((log) => `[${log.timestamp}] [${log.level}] ${log.message}`)
      .join("\n");
    navigator.clipboard.writeText(logText).then(() => {
      updateResultText("Logs copied to clipboard!");
    });
  };

  // Get CSS class for log level
  const getLogLevelClass = (level: string) => {
    switch (level.toUpperCase()) {
      case "DEBUG":
        return "log-level-debug";
      case "INFO":
        return "log-level-info";
      case "WARN":
        return "log-level-warn";
      case "ERROR":
        return "log-level-error";
      default:
        return "log-level-info";
    }
  };

  return (
    <div className={`app-container ${logsPanelOpen ? "logs-open" : ""}`}>
      <button className="close-btn" onClick={() => Quit()}>
        ×
      </button>
      <div className="tabs">
        <button
          className={`tab ${activeTab === "item" ? "active" : ""}`}
          onClick={() => handleTabChange("item")}
        >
          Item
        </button>
        <button
          className={`tab ${activeTab === "order" ? "active" : ""}`}
          onClick={() => handleTabChange("order")}
        >
          Order
        </button>
        <button
          className={`tab ${activeTab === "promotion" ? "active" : ""}`}
          onClick={() => handleTabChange("promotion")}
        >
          Promotion
        </button>
        <button
          className={`tab ${activeTab === "debug" ? "active" : ""}`}
          onClick={() => handleTabChange("debug")}
        >
          Debug
        </button>
      </div>

      {activeTab === "item" && (
        <div className="sub_tabs">
          <button
            className={`tab ${itemSubTab === "create" ? "active" : ""}`}
            onClick={() => handleItemSubTabChange("create")}
          >
            Create
          </button>
          <button
            className={`tab ${itemSubTab === "read" ? "active" : ""}`}
            onClick={() => handleItemSubTabChange("read")}
          >
            Read
          </button>
          <button
            className={`tab ${itemSubTab === "delete" ? "active" : ""}`}
            onClick={() => handleItemSubTabChange("delete")}
          >
            Delete
          </button>
        </div>
      )}

      {activeTab === "order" && (
        <div className="sub_tabs">
          <button
            className={`tab ${orderSubTab === "create" ? "active" : ""}`}
            onClick={() => handleOrderSubTabChange("create")}
          >
            Create
          </button>
          <button
            className={`tab ${orderSubTab === "read" ? "active" : ""}`}
            onClick={() => handleOrderSubTabChange("read")}
          >
            Read
          </button>
          <button
            className={`tab ${orderSubTab === "delete" ? "active" : ""}`}
            onClick={() => handleOrderSubTabChange("delete")}
          >
            Delete
          </button>
        </div>
      )}

      {activeTab === "promotion" && (
        <div className="sub_tabs">
          <button
            className={`tab ${promotionSubTab === "create" ? "active" : ""}`}
            onClick={() => handlePromotionSubTabChange("create")}
          >
            Create
          </button>
          <button
            className={`tab ${promotionSubTab === "read" ? "active" : ""}`}
            onClick={() => handlePromotionSubTabChange("read")}
          >
            Read
          </button>
          <button
            className={`tab ${promotionSubTab === "delete" ? "active" : ""}`}
            onClick={() => handlePromotionSubTabChange("delete")}
          >
            Delete
          </button>
        </div>
      )}

      <div id="App">
        {activeTab !== "order" && activeTab !== "promotion" && (
          <img src={logo} id="logo" alt="logo" />
        )}
        <div id="result" className="result">
          {resultText}
        </div>

        {activeTab === "item" && itemSubTab === "create" && (
          <div id="input" className="input-box">
            <input
              id="name"
              className="input"
              onChange={updateItemText}
              autoComplete="off"
              name="input"
              type="text"
              placeholder="Item Name"
              value={itemText}
            />
            <input
              id="price"
              className="input"
              onChange={updateItemPrice}
              autoComplete="off"
              name="price"
              type="text"
              placeholder="Price ($)"
              value={itemPrice}
            />
            <button className="btn" onClick={addItem}>
              Add Item
            </button>
          </div>
        )}

        {activeTab === "item" && itemSubTab === "read" && (
          <>
            <div id="read-input" className="input-box">
              <input
                id="record-id"
                className="input"
                onChange={updateRecordId}
                autoComplete="off"
                name="record-id"
                placeholder="Enter Record ID"
                value={recordId}
              />
              <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                <label
                  style={{ display: "flex", alignItems: "center", gap: "4px" }}
                >
                  <input
                    type="checkbox"
                    checked={useIndex}
                    onChange={(e: any) => setUseIndex(e.target.checked)}
                  />
                  Use B+ Tree Index
                </label>
              </div>
              <button className="btn" onClick={getRecordById}>
                Get Record
              </button>
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
                <h3 style={{ margin: "0 0 15px 0", color: "#fff" }}>
                  Item Details
                </h3>
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
                    <span style={{ color: "#aaa", fontWeight: "bold" }}>
                      Name:
                    </span>
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
                    <span style={{ color: "#aaa", fontWeight: "bold" }}>
                      Price:
                    </span>
                    <span style={{ color: "#fff" }}>
                      ${(foundItem.priceInCents / 100).toFixed(2)}
                    </span>
                  </div>
                </div>
              </div>
            )}
          </>
        )}

        {activeTab === "item" && itemSubTab === "delete" && (
          <div id="delete-input" className="input-box">
            <input
              id="delete-record-id"
              className="input"
              onChange={updateDeleteRecordId}
              autoComplete="off"
              name="delete-record-id"
              placeholder="Enter Record ID"
              value={deleteRecordId}
            />
            <button className="btn btn-danger" onClick={deleteItem}>
              Delete Record
            </button>
          </div>
        )}

        {activeTab === "order" && orderSubTab === "create" && (
          <div id="order-section">
            <div className="cart-container">
              <div className="cart-header">
                <h3>Order</h3>
                {cart.length > 0 && customerName && (
                  <button className="btn btn-primary" onClick={submitOrder}>
                    Submit Order
                  </button>
                )}
              </div>
              <input
                className="input"
                type="text"
                placeholder="Customer Name"
                value={customerName}
                onChange={updateCustomerName}
                style={{ marginBottom: "10px" }}
              />
              {cart.length > 0 && (
                <div className="cart-total">
                  Total: $
                  {(
                    cart.reduce(
                      (sum, item) => sum + item.priceInCents * item.quantity,
                      0
                    ) / 100
                  ).toFixed(2)}
                </div>
              )}
              <div className="cart-items">
                {cart.length === 0 ? (
                  <div className="cart-empty">Cart is empty</div>
                ) : (
                  cart.map((item) => (
                    <div key={item.id} className="cart-item">
                      <div className="cart-item-info">
                        <div className="cart-item-name">{item.name}</div>
                        <div className="cart-item-id">
                          ID: {item.id} | $
                          {(item.priceInCents / 100).toFixed(2)} each | Total: $
                          {((item.priceInCents * item.quantity) / 100).toFixed(
                            2
                          )}
                        </div>
                      </div>
                      <div className="cart-item-controls">
                        <div className="cart-item-quantity">
                          x{item.quantity}
                        </div>
                        <button
                          className="btn btn-danger btn-small"
                          onClick={() => removeFromCart(item.id)}
                        >
                          ×
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
              <div className="cart-footer">
                <select
                  className="cart-select"
                  value={selectedItemId}
                  onChange={(e: any) => setSelectedItemId(e.target.value)}
                >
                  <option value="">Select an item...</option>
                  {availableItems
                    .sort((a, b) => a.id - b.id)
                    .map((item) => (
                      <option key={item.id} value={item.id}>
                        [{item.id}] {item.name} - $
                        {(item.priceInCents / 100).toFixed(2)}
                      </option>
                    ))}
                </select>
                <button className="btn" onClick={addToCart}>
                  Add
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === "order" && orderSubTab === "read" && (
          <div id="order-read-input" className="input-box">
            <input
              id="order-read-id"
              className="input"
              onChange={updateOrderReadId}
              autoComplete="off"
              name="order-read-id"
              placeholder="Enter Order ID"
              value={orderReadId}
            />
            <button className="btn" onClick={getOrderById}>
              Get Order
            </button>
          </div>
        )}

        {activeTab === "order" && orderSubTab === "delete" && (
          <div id="order-delete-input" className="input-box">
            <input
              id="order-delete-id"
              className="input"
              onChange={updateOrderDeleteId}
              autoComplete="off"
              name="order-delete-id"
              placeholder="Enter Order ID"
              value={orderDeleteId}
            />
            <button className="btn btn-danger" onClick={deleteOrderById}>
              Delete Order
            </button>
          </div>
        )}

        {activeTab === "promotion" && promotionSubTab === "create" && (
          <div id="promotion-section">
            <div className="cart-container">
              <div className="cart-header">
                <h3>Promotion</h3>
                {promotionCart.length > 0 && promotionName && (
                  <button className="btn btn-primary" onClick={submitPromotion}>
                    Create Promotion
                  </button>
                )}
              </div>
              <input
                className="input"
                type="text"
                placeholder="Promotion Name"
                value={promotionName}
                onChange={updatePromotionName}
                style={{ marginBottom: "10px" }}
              />
              {promotionCart.length > 0 && (
                <div className="cart-total">
                  Total: $
                  {(
                    promotionCart.reduce(
                      (sum, item) => sum + item.priceInCents * item.quantity,
                      0
                    ) / 100
                  ).toFixed(2)}
                </div>
              )}
              <div className="cart-items">
                {promotionCart.length === 0 ? (
                  <div className="cart-empty">No items in promotion</div>
                ) : (
                  promotionCart.map((item) => (
                    <div key={item.id} className="cart-item">
                      <div className="cart-item-info">
                        <div className="cart-item-name">{item.name}</div>
                        <div className="cart-item-id">
                          ID: {item.id} | $
                          {(item.priceInCents / 100).toFixed(2)} each | Total: $
                          {((item.priceInCents * item.quantity) / 100).toFixed(
                            2
                          )}
                        </div>
                      </div>
                      <div className="cart-item-controls">
                        <div className="cart-item-quantity">
                          x{item.quantity}
                        </div>
                        <button
                          className="btn btn-danger btn-small"
                          onClick={() => removeFromPromotionCart(item.id)}
                        >
                          ×
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
              <div className="cart-footer">
                <select
                  className="cart-select"
                  value={selectedPromotionItemId}
                  onChange={(e: any) => setSelectedPromotionItemId(e.target.value)}
                >
                  <option value="">Select an item...</option>
                  {availableItems
                    .sort((a, b) => a.id - b.id)
                    .map((item) => (
                      <option key={item.id} value={item.id}>
                        [{item.id}] {item.name} - $
                        {(item.priceInCents / 100).toFixed(2)}
                      </option>
                    ))}
                </select>
                <button className="btn" onClick={addToPromotionCart}>
                  Add
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === "promotion" && promotionSubTab === "read" && (
          <div id="promotion-read-input" className="input-box">
            <input
              id="promotion-read-id"
              className="input"
              onChange={updatePromotionReadId}
              autoComplete="off"
              name="promotion-read-id"
              placeholder="Enter Promotion ID"
              value={promotionReadId}
            />
            <button className="btn" onClick={getPromotionById}>
              Get Promotion
            </button>
          </div>
        )}

        {activeTab === "promotion" && promotionSubTab === "delete" && (
          <div id="promotion-delete-input" className="input-box">
            <input
              id="promotion-delete-id"
              className="input"
              onChange={updatePromotionDeleteId}
              autoComplete="off"
              name="promotion-delete-id"
              placeholder="Enter Promotion ID"
              value={promotionDeleteId}
            />
            <button className="btn btn-danger" onClick={deletePromotionById}>
              Delete Promotion
            </button>
          </div>
        )}

        {activeTab === "debug" && (
          <div id="debug-controls" className="debug-section">
            <div className="input-box">
              <button className="btn" onClick={populateInventory}>
                Populate Inventory
              </button>
              <button className="btn" onClick={printIndex}>
                Print Index
              </button>
              <button className="btn btn-danger" onClick={deleteAllFiles}>
                Delete All Files
              </button>
            </div>

            {indexData && (
              <div
                style={{
                  marginTop: "20px",
                  padding: "20px",
                  backgroundColor: "rgba(255, 255, 255, 0.05)",
                  borderRadius: "8px",
                  border: "1px solid rgba(255, 255, 255, 0.1)",
                  maxHeight: "400px",
                  overflowY: "auto",
                }}
              >
                <h3 style={{ margin: "0 0 15px 0", color: "#fff" }}>
                  B+ Tree Index Contents
                </h3>
                <div style={{ marginBottom: "10px", color: "#aaa" }}>
                  Total entries: {indexData.count}
                </div>
                <table
                  style={{
                    width: "100%",
                    borderCollapse: "collapse",
                    color: "#fff",
                  }}
                >
                  <thead>
                    <tr
                      style={{
                        borderBottom: "2px solid rgba(255, 255, 255, 0.2)",
                      }}
                    >
                      <th style={{ padding: "8px", textAlign: "left" }}>
                        Item ID
                      </th>
                      <th style={{ padding: "8px", textAlign: "left" }}>
                        File Offset
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {indexData.entries.map((entry: any, idx: number) => (
                      <tr
                        key={idx}
                        style={{
                          borderBottom: "1px solid rgba(255, 255, 255, 0.1)",
                          backgroundColor:
                            idx % 2 === 0
                              ? "rgba(0, 0, 0, 0.2)"
                              : "transparent",
                        }}
                      >
                        <td style={{ padding: "8px" }}>{entry.id}</td>
                        <td
                          style={{
                            padding: "8px",
                            fontFamily: "monospace",
                            color: "#888",
                          }}
                        >
                          {entry.offset} bytes
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Logs Side Panel */}
      <div className={`logs-panel ${logsPanelOpen ? "open" : "closed"}`}>
        <button
          className="logs-toggle"
          onClick={() => setLogsPanelOpen(!logsPanelOpen)}
        >
          {logsPanelOpen ? "»" : "«"}
        </button>

        {logsPanelOpen && (
          <div className="logs-content">
            <div className="logs-header">
              <h3>Logs</h3>
              <div className="logs-controls">
                <button
                  className="btn-icon"
                  onClick={copyLogs}
                  title="Copy Logs"
                >
                  📋
                </button>
                <button
                  className="btn-icon btn-danger"
                  onClick={clearLogs}
                  title="Clear Logs"
                >
                  🗑
                </button>
              </div>
            </div>
            <div className="logs-container" ref={logsContainerRef}>
              {logs.length === 0 ? (
                <div className="logs-empty">No logs yet</div>
              ) : (
                logs.map((log, index) => (
                  <div key={index} className={`log-entry ${getLogLevelClass(log.level)}`}>
                    <span className="log-timestamp">[{log.timestamp}]</span>
                    <span className="log-level">[{log.level}]</span>
                    <span className="log-message">{log.message}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

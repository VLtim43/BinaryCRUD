import "./App.scss";
import logo from "./assets/images/logo-universal.png";
import {
  AddItem,
  AddOrder,
  AddPromotion,
  PrintBinaryFile,
  PrintOrdersFile,
  PrintPromotionsFile,
  DeleteAllFiles,
  GetItemByID,
  GetItemByIDWithIndex,
  GetOrderByID,
  GetPromotionByID,
  GetItems,
  GetOrders,
  GetPromotions,
  PopulateInventory,
  DeleteItem,
  PrintIndex,
  RebuildIndex,
} from "../wailsjs/go/main/App";
import { Quit } from "../wailsjs/runtime/runtime";
import { useState } from "preact/hooks";
import { h, Fragment } from "preact";

export const App = () => {
  const [activeTab, setActiveTab] = useState<"item" | "order" | "promotion" | "debug">("item");
  const [itemSubTab, setItemSubTab] = useState<"create" | "read" | "delete">("create");
  const [orderSubTab, setOrderSubTab] = useState<"create" | "read" | "delete">("create");
  const [promotionSubTab, setPromotionSubTab] = useState<"create" | "read" | "delete">("create");
  const [resultText, setResultText] = useState("Enter item text below 👇");
  const [itemText, setItemText] = useState("");
  const [itemPrice, setItemPrice] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteRecordId, setDeleteRecordId] = useState("");
  const [availableItems, setAvailableItems] = useState<Array<{id: number, name: string, priceInCents: number}>>([]);
  const [cart, setCart] = useState<Array<{id: number, name: string, quantity: number, priceInCents: number}>>([]);
  const [selectedItemId, setSelectedItemId] = useState<string>("");
  const [promotionName, setPromotionName] = useState("");
  const [promotionCart, setPromotionCart] = useState<Array<{id: number, name: string, quantity: number, priceInCents: number}>>([]);
  const [promotionSelectedItemId, setPromotionSelectedItemId] = useState<string>("");
  const [orderReadId, setOrderReadId] = useState("");
  const [orderDeleteId, setOrderDeleteId] = useState("");
  const [promotionReadId, setPromotionReadId] = useState("");
  const [promotionDeleteId, setPromotionDeleteId] = useState("");
  const [useIndex, setUseIndex] = useState(true);
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

    if (tab === "promotion") {
      switch (subTab) {
        case "create":
          return "Create a promotion with a name and items";
        case "read":
          return "Enter a promotion ID to fetch 👇";
        case "delete":
          return "Enter a promotion ID to delete 👇";
        default:
          return "Create a promotion with a name and items";
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
      setResultText(getDefaultText(tab, itemSubTab));
    } else if (tab === "order") {
      setResultText(getDefaultText(tab, orderSubTab));
      // Load items when entering order tab and subtab is create
      if (orderSubTab === "create") {
        loadItems();
      }
    } else if (tab === "promotion") {
      setResultText(getDefaultText(tab, promotionSubTab));
      // Load items when entering promotion tab and subtab is create
      if (promotionSubTab === "create") {
        loadItems();
      }
    } else {
      setResultText(getDefaultText(tab));
    }
  };

  // Handle item subtab changes
  const handleItemSubTabChange = (subTab: "create" | "read" | "delete") => {
    setItemSubTab(subTab);
    setResultText(getDefaultText("item", subTab));
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
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const printFile = () => {
    PrintBinaryFile()
      .then(() => {
        updateResultText("Binary file printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const getRecordById = () => {
    // Validate input before sending to backend
    if (!recordId || recordId.trim().length === 0) {
      updateResultText("Error: Please enter a record ID");
      return;
    }

    // Parse the record ID as a number
    const id = parseInt(recordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Record ID must be a valid non-negative number");
      return;
    }

    // Use index or sequential search based on checkbox
    const searchMethod = useIndex ? GetItemByIDWithIndex : GetItemByID;
    const methodName = useIndex ? "B+ Tree Index" : "Sequential Search";

    searchMethod(id)
      .then((item: any) => {
        const price = (item.priceInCents / 100).toFixed(2);
        updateResultText(`Record ${id}: ${item.name} - $${price} (using ${methodName})`);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const deleteAllFiles = () => {
    DeleteAllFiles()
      .then(() => {
        updateResultText("All generated files deleted successfully!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const populateInventory = () => {
    PopulateInventory("inventory.json")
      .then((result: string) => {
        updateResultText(`Inventory populated! ${result}`);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const printIndex = () => {
    PrintIndex()
      .then(() => {
        updateResultText("Index printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const rebuildIndex = () => {
    RebuildIndex()
      .then(() => {
        updateResultText("Index rebuilt successfully!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const deleteItem = () => {
    // Validate input before sending to backend
    if (!deleteRecordId || deleteRecordId.trim().length === 0) {
      updateResultText("Error: Please enter a record ID");
      return;
    }

    // Parse the record ID as a number
    const id = parseInt(deleteRecordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Record ID must be a valid non-negative number");
      return;
    }

    DeleteItem(id)
      .then((itemName: string) => {
        updateResultText(`Record [${id}] [${itemName}] was deleted`);
        setDeleteRecordId("");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Load all items for order tab
  const loadItems = () => {
    GetItems()
      .then((items: Array<{id: number, name: string, priceInCents: number}>) => {
        setAvailableItems(items);
      })
      .catch((err: any) => {
        updateResultText(`Error loading items: ${err}`);
      });
  };

  // Add item to cart
  const addToCart = () => {
    if (!selectedItemId) {
      updateResultText("Error: Please select an item");
      return;
    }

    const itemId = parseInt(selectedItemId, 10);
    const item = availableItems.find(i => i.id === itemId);

    if (!item) {
      updateResultText("Error: Item not found");
      return;
    }

    // Check if item already in cart
    const existingItem = cart.find(c => c.id === itemId);
    if (existingItem) {
      // Increment quantity
      setCart(cart.map(c =>
        c.id === itemId ? { ...c, quantity: c.quantity + 1 } : c
      ));
    } else {
      // Add new item to cart
      setCart([...cart, { id: item.id, name: item.name, quantity: 1, priceInCents: item.priceInCents }]);
    }

    updateResultText(`Added ${item.name} to cart`);
  };

  // Remove item from cart
  const removeFromCart = (itemId: number) => {
    const item = cart.find(c => c.id === itemId);
    if (item) {
      setCart(cart.filter(c => c.id !== itemId));
      updateResultText(`Removed ${item.name} from cart`);
    }
  };

  // Submit order - writes order to orders.bin
  const submitOrder = () => {
    if (cart.length === 0) {
      updateResultText("Error: Cart is empty");
      return;
    }

    // Build array of item names based on quantities
    const itemNames: string[] = [];
    cart.forEach(item => {
      for (let i = 0; i < item.quantity; i++) {
        itemNames.push(item.name);
      }
    });

    AddOrder(itemNames)
      .then(() => {
        updateResultText(`Order submitted with ${cart.length} unique item(s) (${itemNames.length} total)!`);
        setCart([]);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Add item to promotion cart
  const addToPromotionCart = () => {
    if (!promotionSelectedItemId) {
      updateResultText("Error: Please select an item");
      return;
    }

    const itemId = parseInt(promotionSelectedItemId, 10);
    const item = availableItems.find(i => i.id === itemId);

    if (!item) {
      updateResultText("Error: Item not found");
      return;
    }

    // Check if item already in promotion cart
    const existingItem = promotionCart.find(c => c.id === itemId);
    if (existingItem) {
      // Increment quantity
      setPromotionCart(promotionCart.map(c =>
        c.id === itemId ? { ...c, quantity: c.quantity + 1 } : c
      ));
    } else {
      // Add new item to promotion cart
      setPromotionCart([...promotionCart, { id: item.id, name: item.name, quantity: 1, priceInCents: item.priceInCents }]);
    }

    updateResultText(`Added ${item.name} to promotion`);
  };

  // Remove item from promotion cart
  const removeFromPromotionCart = (itemId: number) => {
    const item = promotionCart.find(c => c.id === itemId);
    if (item) {
      setPromotionCart(promotionCart.filter(c => c.id !== itemId));
      updateResultText(`Removed ${item.name} from promotion`);
    }
  };

  // Submit promotion - writes promotion to promotions.bin
  const submitPromotion = () => {
    if (!promotionName || promotionName.trim().length === 0) {
      updateResultText("Error: Please enter a promotion name");
      return;
    }

    if (promotionCart.length === 0) {
      updateResultText("Error: Promotion cart is empty");
      return;
    }

    // Build array of item names based on quantities
    const itemNames: string[] = [];
    promotionCart.forEach(item => {
      for (let i = 0; i < item.quantity; i++) {
        itemNames.push(item.name);
      }
    });

    AddPromotion(promotionName, itemNames)
      .then(() => {
        updateResultText(`Promotion "${promotionName}" created with ${promotionCart.length} unique item(s) (${itemNames.length} total)!`);
        setPromotionName("");
        setPromotionCart([]);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
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
      updateResultText("Error: Order ID must be a valid non-negative number");
      return;
    }

    GetOrderByID(id)
      .then((order: any) => {
        const itemsList = order.items.join(", ");
        updateResultText(`Order ${id}: ${order.items.length} item(s) - ${itemsList}`);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Print orders file
  const printOrdersFile = () => {
    PrintOrdersFile()
      .then(() => {
        updateResultText("Orders file printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Delete order placeholder (not implemented yet)
  const deleteOrder = () => {
    updateResultText("Delete order functionality not yet implemented");
  };

  // Get promotion by ID
  const getPromotionById = () => {
    if (!promotionReadId || promotionReadId.trim().length === 0) {
      updateResultText("Error: Please enter a promotion ID");
      return;
    }

    const id = parseInt(promotionReadId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Promotion ID must be a valid non-negative number");
      return;
    }

    GetPromotionByID(id)
      .then((promotion: any) => {
        const itemsList = promotion.items.join(", ");
        updateResultText(`Promotion ${id} "${promotion.name}": ${promotion.items.length} item(s) - ${itemsList}`);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Print promotions file
  const printPromotionsFile = () => {
    PrintPromotionsFile()
      .then(() => {
        updateResultText("Promotions file printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  // Delete promotion placeholder (not implemented yet)
  const deletePromotion = () => {
    updateResultText("Delete promotion functionality not yet implemented");
  };

  return (
    <>
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
        {activeTab !== "order" && activeTab !== "promotion" && <img src={logo} id="logo" alt="logo" />}
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
              <label style={{ display: "flex", alignItems: "center", gap: "4px" }}>
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
            <button className="btn" onClick={printFile}>
              Print
            </button>
          </div>
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
                <h3>Cart</h3>
                {cart.length > 0 && (
                  <button className="btn btn-primary" onClick={submitOrder}>
                    Submit Order
                  </button>
                )}
              </div>

              {cart.length > 0 && (
                <div className="cart-total">
                  Total: ${(cart.reduce((sum, item) => sum + (item.priceInCents * item.quantity), 0) / 100).toFixed(2)}
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
                        <div className="cart-item-id">ID: {item.id} | ${(item.priceInCents / 100).toFixed(2)}</div>
                      </div>
                      <div className="cart-item-controls">
                        <div className="cart-item-quantity">x{item.quantity}</div>
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
                  {availableItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      [{item.id}] {item.name} - ${(item.priceInCents / 100).toFixed(2)}
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
            <button className="btn" onClick={printOrdersFile}>
              Print
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
            <button className="btn btn-danger" onClick={deleteOrder}>
              Delete Order
            </button>
          </div>
        )}

        {activeTab === "promotion" && promotionSubTab === "create" && (
          <div id="promotion-section">
            <div className="cart-container">
              <div className="cart-header">
                <h3>Create Promotion</h3>
              </div>

              <div className="promotion-name-input">
                <input
                  className="input"
                  placeholder="Promotion Name"
                  value={promotionName}
                  onChange={(e: any) => setPromotionName(e.target.value)}
                  autoComplete="off"
                />
              </div>

              {promotionCart.length > 0 && (
                <div className="cart-total">
                  Total: ${(promotionCart.reduce((sum, item) => sum + (item.priceInCents * item.quantity), 0) / 100).toFixed(2)}
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
                        <div className="cart-item-id">ID: {item.id} | ${(item.priceInCents / 100).toFixed(2)}</div>
                      </div>
                      <div className="cart-item-controls">
                        <div className="cart-item-quantity">x{item.quantity}</div>
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
                  value={promotionSelectedItemId}
                  onChange={(e: any) => setPromotionSelectedItemId(e.target.value)}
                >
                  <option value="">Select an item...</option>
                  {availableItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      [{item.id}] {item.name} - ${(item.priceInCents / 100).toFixed(2)}
                    </option>
                  ))}
                </select>
                <button className="btn" onClick={addToPromotionCart}>
                  Add
                </button>
              </div>

              {promotionName && promotionCart.length > 0 && (
                <div className="promotion-submit">
                  <button className="btn btn-primary" onClick={submitPromotion}>
                    Create Promotion
                  </button>
                </div>
              )}
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
            <button className="btn" onClick={printPromotionsFile}>
              Print
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
            <button className="btn btn-danger" onClick={deletePromotion}>
              Delete Promotion
            </button>
          </div>
        )}

        {activeTab === "debug" && (
          <div id="debug-controls" className="debug-section">
            <div className="input-box">
              <button className="btn btn-warning" onClick={populateInventory}>
                Populate Inventory
              </button>
              <button className="btn" onClick={printIndex}>
                Print Index
              </button>
              <button className="btn" onClick={rebuildIndex}>
                Rebuild Index
              </button>
              <button className="btn btn-danger" onClick={deleteAllFiles}>
                Delete All Files
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
};

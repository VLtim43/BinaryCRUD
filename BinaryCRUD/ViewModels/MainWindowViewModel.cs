using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Threading.Tasks;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Controls.ApplicationLifetimes;
using BinaryCRUD.Models;
using BinaryCRUD.Services;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace BinaryCRUD.ViewModels;

public partial class MainWindowViewModel : ViewModelBase
{
    private readonly ItemDAO _itemDAO;
    private readonly OrderDAO _orderDAO;
    public ToastService ToastService { get; }
    public ConsoleService ConsoleService { get; }
    public AuthenticationService AuthService { get; }

    public MainWindowViewModel(ItemDAO itemDAO)
    {
        _itemDAO = itemDAO;
        _orderDAO = new OrderDAO();
        ToastService = new ToastService();
        ConsoleService = new ConsoleService();
        AuthService = new AuthenticationService();
        AuthService.UserChanged += OnUserChanged;
        _ = InitializeAsync();
    }

    private async Task InitializeAsync()
    {
        await AuthService.InitializeAsync();
        await LoadItemsAsync();
        await LoadOrdersAsync();
    }

    private void OnUserChanged(object? sender, User? user)
    {
        OnPropertyChanged(nameof(CurrentUser));
        OnPropertyChanged(nameof(IsLoggedIn));
        OnPropertyChanged(nameof(IsAdmin));
        OnPropertyChanged(nameof(CanCreateItems));
        OnPropertyChanged(nameof(CanDeleteItems));
        OnPropertyChanged(nameof(CanManageInventory));
        OnPropertyChanged(nameof(CanDeleteFiles));
        OnPropertyChanged(nameof(CanCreateOrders));
        OnPropertyChanged(nameof(CanDeleteOrders));
        OnPropertyChanged(nameof(CanViewOrders));
        OnPropertyChanged(nameof(LoginButtonText));
    }

    [ObservableProperty]
    private string text = string.Empty;

    [ObservableProperty]
    private decimal priceInput = 0.0m;

    [ObservableProperty]
    private ObservableCollection<Item> items = new();

    [ObservableProperty]
    private bool hasItems = false;

    [ObservableProperty]
    private ObservableCollection<Order> orders = new();

    [ObservableProperty]
    private bool hasOrders = false;

    [ObservableProperty]
    private bool hasUsers = false;

    [ObservableProperty]
    private Item? selectedItem = null;

    [ObservableProperty]
    private ObservableCollection<Item> activeItems = new();

    [ObservableProperty]
    private bool isConsoleVisible = false;

    [ObservableProperty]
    private ObservableCollection<ushort> cartItemIds = new();

    [ObservableProperty]
    private decimal cartTotal = 0.0m;

    [ObservableProperty]
    private string? orderAdditionalInfo = null;

    // Login properties
    [ObservableProperty]
    private string loginUsername = string.Empty;

    [ObservableProperty]
    private string loginPassword = string.Empty;

    // User management properties
    [ObservableProperty]
    private string newUsername = string.Empty;

    [ObservableProperty]
    private string newPassword = string.Empty;

    [ObservableProperty]
    private int newUserRoleIndex = 0; // 0 = User, 1 = Admin

    [ObservableProperty]
    private ObservableCollection<User> users = new();

    // Authentication properties
    public User? CurrentUser => AuthService.CurrentUser;
    public bool IsLoggedIn => AuthService.IsLoggedIn;
    public bool IsAdmin => AuthService.IsAdmin;
    public string LoginButtonText => IsLoggedIn ? $"Logout ({CurrentUser?.Username})" : "Login";

    // Permission properties for UI binding
    public bool CanCreateItems => AuthService.HasPermission("create_items");
    public bool CanDeleteItems => AuthService.HasPermission("delete_items");
    public bool CanManageInventory => AuthService.HasPermission("populate_inventory");
    public bool CanDeleteFiles => AuthService.HasPermission("delete_files");
    public bool CanCreateOrders => AuthService.HasPermission("create_orders");
    public bool CanDeleteOrders => AuthService.HasPermission("delete_orders");
    public bool CanViewOrders => AuthService.HasPermission("view_orders");

    public IEnumerable<CartDisplayItem> CartDisplayItems
    {
        get
        {
            return CartItemIds.Select(id => new CartDisplayItem
            {
                ItemId = id,
                ItemName = Items.FirstOrDefault(i => i.Id == id)?.Content ?? "Unknown Item",
            });
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task LoginAsync()
    {
        if (string.IsNullOrWhiteSpace(LoginUsername) || string.IsNullOrWhiteSpace(LoginPassword))
        {
            ToastService.ShowWarning("Please enter both username and password");
            return;
        }

        var success = await AuthService.LoginAsync(LoginUsername, LoginPassword);

        if (success)
        {
            ToastService.ShowSuccess($"Welcome, {CurrentUser?.Username}!");
            LoginUsername = string.Empty;
            LoginPassword = string.Empty;

            // Load users for admin
            if (IsAdmin)
            {
                await LoadUsersAsync();
            }
        }
        else
        {
            ToastService.ShowWarning("Invalid username or password");
            LoginPassword = string.Empty;
        }
    }

    [RelayCommand]
    private void Logout()
    {
        AuthService.Logout();
        ToastService.ShowSuccess("Logged out successfully");

        // Clear sensitive data
        LoginUsername = string.Empty;
        LoginPassword = string.Empty;
        Users.Clear();
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task SaveAsync()
    {
        if (!AuthService.HasPermission("create_items"))
        {
            ToastService.ShowWarning("You don't have permission to create items");
            return;
        }

        if (string.IsNullOrEmpty(Text))
        {
            System.Console.WriteLine("[WARNING] Cannot save empty item");
            ToastService.ShowWarning("Cannot save empty item");
            return;
        }

        try
        {
            System.Console.WriteLine($"[INFO] Creating item: '{Text}' with price: ${PriceInput}");
            await _itemDAO.AddItemAsync(Text, PriceInput);
            ToastService.ShowSuccess($"Item '{Text}' (${PriceInput:F2}) created successfully");
            Text = string.Empty;
            PriceInput = 0.0m;
            await LoadItemsAsync();
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to create item: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task ReadHeaderAsync()
    {
        try
        {
            System.Console.WriteLine("[INFO] Reading file header...");
            var header = await _itemDAO.ReadHeaderAsync();

            if (header == null)
            {
                System.Console.WriteLine("[INFO] No file found or file is empty");
                ToastService.ShowSuccess("No file found");
                return;
            }

            System.Console.WriteLine($"[HEADER] Count: {header.Count} items");
            System.Console.WriteLine($"[HEADER] Header Size: 4 bytes");
            System.Console.WriteLine($"[HEADER] File Path: item.bin");

            ToastService.ShowSuccess($"Header: {header.Count} items");
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to read header: {ex.Message}");
            ToastService.ShowSuccess("Error reading header");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task ReadItemsAsync()
    {
        try
        {
            System.Console.WriteLine("[INFO] Reading items from file...");
            var items = await _itemDAO.GetAllItemsAsync();

            if (items.Count == 0)
            {
                System.Console.WriteLine("[INFO] No items found");
                return;
            }

            // Console output: Show all items in chronological order (oldest first)
            var orderedItems = items.OrderBy(x => x.Id).ToList();
            for (int i = 0; i < orderedItems.Count; i++)
            {
                var item = orderedItems[i];
                var status = item.IsTombstone ? "DELETED" : "ACTIVE";
                System.Console.WriteLine(
                    $"[ID: {item.Id}][{status}] - [Name: '{item.Content}'] - [Price: ${item.Price:F2}]"
                );
            }

            System.Console.WriteLine($"[INFO] Found {items.Count} items total");
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to read items: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task RefreshItemsAsync()
    {
        await LoadItemsAsync();
        ToastService.ShowSuccess("Items list refreshed");
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task DeleteItemAsync(ushort itemId)
    {
        try
        {
            System.Console.WriteLine($"[INFO] Deleting item with ID: {itemId}");
            await _itemDAO.DeleteItemAsync(itemId);
            ToastService.ShowSuccess($"Item {itemId} marked as deleted");
            await LoadItemsAsync();
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to delete item: {ex.Message}");
            ToastService.ShowWarning($"Failed to delete item: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task CreateOrderAsync()
    {
        if (SelectedItem == null)
        {
            System.Console.WriteLine("[WARNING] Cannot create order without selecting an item");
            ToastService.ShowWarning("Please select an item to create an order");
            return;
        }

        if (SelectedItem.IsTombstone)
        {
            System.Console.WriteLine("[WARNING] Cannot create order for deleted item");
            ToastService.ShowWarning("Cannot create order for deleted item");
            return;
        }

        try
        {
            if (CurrentUser == null)
            {
                System.Console.WriteLine("[WARNING] Cannot create order without logged in user");
                ToastService.ShowWarning("You must be logged in to create orders");
                return;
            }

            System.Console.WriteLine(
                $"[INFO] Creating order for item ID: {SelectedItem.Id} by user {CurrentUser.Username}"
            );
            await _orderDAO.AddOrderAsync(SelectedItem.Id, SelectedItem.Price, CurrentUser.Id);
            ToastService.ShowSuccess(
                $"Order created for '{SelectedItem.Content}' (${SelectedItem.Price:F2})"
            );
            await LoadOrdersAsync();
            SelectedItem = null;
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to create order: {ex.Message}");
            ToastService.ShowWarning($"Failed to create order: {ex.Message}");
        }
    }

    [RelayCommand]
    private void AddToCart()
    {
        if (SelectedItem == null)
        {
            System.Console.WriteLine("[WARNING] Cannot add to cart without selecting an item");
            ToastService.ShowWarning("Please select an item to add to cart");
            return;
        }

        if (SelectedItem.IsTombstone)
        {
            System.Console.WriteLine("[WARNING] Cannot add deleted item to cart");
            ToastService.ShowWarning("Cannot add deleted item to cart");
            return;
        }

        CartItemIds.Add(SelectedItem.Id);
        System.Console.WriteLine(
            $"[INFO] Added item ID {SelectedItem.Id} to cart (total items: {CartItemIds.Count})"
        );

        UpdateCartTotal();
        OnPropertyChanged(nameof(CartDisplayItems));
        ToastService.ShowSuccess($"Added '{SelectedItem.Content}' to cart");
        SelectedItem = null;
    }

    [RelayCommand]
    private void RemoveFromCart(ushort itemId)
    {
        CartItemIds.Remove(itemId);
        UpdateCartTotal();
        OnPropertyChanged(nameof(CartDisplayItems));
        System.Console.WriteLine(
            $"[INFO] Removed item ID {itemId} from cart (total items: {CartItemIds.Count})"
        );
        ToastService.ShowSuccess("Item removed from cart");
    }

    [RelayCommand]
    private void ClearCart()
    {
        var itemCount = CartItemIds.Count;
        CartItemIds.Clear();
        UpdateCartTotal();
        OnPropertyChanged(nameof(CartDisplayItems));
        System.Console.WriteLine($"[INFO] Cleared cart ({itemCount} items removed)");
        ToastService.ShowSuccess("Cart cleared");
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task CreateOrderFromCartAsync()
    {
        if (!AuthService.HasPermission("create_orders"))
        {
            ToastService.ShowWarning("You need to be logged in to create orders");
            return;
        }

        if (CartItemIds.Count == 0)
        {
            System.Console.WriteLine("[WARNING] Cannot create order with empty cart");
            ToastService.ShowWarning("Please add items to cart before creating order");
            return;
        }

        try
        {
            var totalPrice = (float)CartTotal;

            if (CurrentUser == null)
            {
                System.Console.WriteLine("[WARNING] Cannot create order without logged in user");
                ToastService.ShowWarning("You must be logged in to create orders");
                return;
            }

            System.Console.WriteLine(
                $"[INFO] Creating order for user '{CurrentUser.Username}' with {CartItemIds.Count} items, total: ${totalPrice:F2}"
            );
            await _orderDAO.AddOrderAsync(
                CartItemIds.ToList(),
                totalPrice,
                CurrentUser.Id,
                string.IsNullOrWhiteSpace(OrderAdditionalInfo) ? null : OrderAdditionalInfo
            );

            ToastService.ShowSuccess(
                $"Order for '{CurrentUser.Username}' created successfully (${totalPrice:F2})"
            );

            await LoadOrdersAsync();
            ClearOrderForm();
            ClearCart();
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to create order: {ex.Message}");
            ToastService.ShowWarning($"Failed to create order: {ex.Message}");
        }
    }

    private void ClearOrderForm()
    {
        OrderAdditionalInfo = null;
    }

    private void UpdateCartTotal()
    {
        decimal total = 0;
        foreach (var itemId in CartItemIds)
        {
            var item = Items.FirstOrDefault(i => i.Id == itemId);
            if (item != null)
            {
                total += (decimal)item.Price;
            }
        }
        CartTotal = total;
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task DeleteOrderAsync(ushort orderId)
    {
        try
        {
            System.Console.WriteLine($"[INFO] Deleting order with ID: {orderId}");
            await _orderDAO.DeleteOrderAsync(orderId);
            ToastService.ShowSuccess($"Order {orderId} marked as deleted");
            await LoadOrdersAsync();
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to delete order: {ex.Message}");
            ToastService.ShowWarning($"Failed to delete order: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task RefreshOrdersAsync()
    {
        await LoadOrdersAsync();
        ToastService.ShowSuccess("Orders list refreshed");
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task ReadOrdersAsync()
    {
        try
        {
            System.Console.WriteLine("[INFO] Reading orders from file...");
            var orders = await _orderDAO.GetAllOrdersAsync();

            if (orders.Count == 0)
            {
                System.Console.WriteLine("[INFO] No orders found");
                return;
            }

            var orderedOrders = orders.OrderBy(x => x.Id).ToList();
            for (int i = 0; i < orderedOrders.Count; i++)
            {
                var order = orderedOrders[i];
                var status = order.IsTombstone ? "DELETED" : "ACTIVE";
                var itemsDisplay = string.Join(", ", order.ItemIds.Select(id => $"ID:{id}"));
                var additionalInfoDisplay = !string.IsNullOrEmpty(order.AdditionalInfo)
                    ? $" - [Info: {order.AdditionalInfo}]"
                    : "";
                System.Console.WriteLine(
                    $"[ID: {order.Id}][{status}] - [UserId: {order.UserId}] - [Items: {itemsDisplay}] - [Total: ${order.TotalPrice:F2}]{additionalInfoDisplay}"
                );
            }

            System.Console.WriteLine($"[INFO] Found {orders.Count} orders total");
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to read orders: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task DeleteAllOrdersAsync()
    {
        try
        {
            var result = await ShowConfirmationDialogAsync(
                "Delete File",
                "Are you sure you want to delete the entire order.bin file?\n\nThis will permanently remove all orders and cannot be undone."
            );

            if (result)
            {
                await _orderDAO.DeleteFileAsync();
                Orders.Clear();
                HasOrders = false;
                ToastService.ShowWarning("Orders file deleted successfully");
                System.Console.WriteLine("[INFO] order.bin file deleted");
            }
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to delete orders file: {ex.Message}");
            ToastService.ShowWarning($"Error deleting orders file: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task DeleteFileAsync()
    {
        try
        {
            var result = await ShowConfirmationDialogAsync(
                "Delete File",
                "Are you sure you want to delete the entire item.bin file?\n\nThis will permanently remove all items and cannot be undone."
            );

            if (result)
            {
                await _itemDAO.DeleteFileAsync();
                Items.Clear();
                HasItems = false;
                ToastService.ShowWarning("File deleted successfully");
                System.Console.WriteLine("[INFO] item.bin file deleted");
            }
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to delete file: {ex.Message}");
            ToastService.ShowSuccess($"Error deleting file: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task DeleteUsersFileAsync()
    {
        try
        {
            var result = await ShowConfirmationDialogAsync(
                "Delete File",
                "Are you sure you want to delete the entire users.bin file?\n\nThis will permanently remove all users and cannot be undone."
            );

            if (result)
            {
                var userDAO = new UserDAO();
                await userDAO.DeleteFileAsync();
                Users.Clear();
                HasUsers = false;
                ToastService.ShowWarning("Users file deleted successfully");
                System.Console.WriteLine("[INFO] users.bin file deleted");
            }
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to delete users file: {ex.Message}");
            ToastService.ShowSuccess($"Error deleting users file: {ex.Message}");
        }
    }

    private async Task<bool> ShowConfirmationDialogAsync(string title, string message)
    {
        if (
            Application.Current?.ApplicationLifetime
            is IClassicDesktopStyleApplicationLifetime desktop
        )
        {
            var mainWindow = desktop.MainWindow;
            if (mainWindow != null)
            {
                bool? result = null;

                var yesButton = new Button
                {
                    Content = "Yes",
                    Width = 80,
                    Height = 35,
                };

                var noButton = new Button
                {
                    Content = "No",
                    Width = 80,
                    Height = 35,
                };

                var dialog = new Window
                {
                    Title = title,
                    Width = 400,
                    Height = 200,
                    WindowStartupLocation = WindowStartupLocation.CenterOwner,
                    CanResize = false,
                    Content = new StackPanel
                    {
                        Margin = new Avalonia.Thickness(20),
                        Spacing = 15,
                        Children =
                        {
                            new TextBlock
                            {
                                Text = message,
                                TextWrapping = Avalonia.Media.TextWrapping.Wrap,
                                FontSize = 14,
                            },
                            new StackPanel
                            {
                                Orientation = Avalonia.Layout.Orientation.Horizontal,
                                HorizontalAlignment = Avalonia.Layout.HorizontalAlignment.Center,
                                Spacing = 10,
                                Children = { yesButton, noButton },
                            },
                        },
                    },
                };

                yesButton.Click += (s, e) =>
                {
                    result = true;
                    dialog.Close();
                };
                noButton.Click += (s, e) =>
                {
                    result = false;
                    dialog.Close();
                };

                await dialog.ShowDialog(mainWindow);
                return result == true;
            }
        }
        return false;
    }

    private async System.Threading.Tasks.Task LoadItemsAsync()
    {
        try
        {
            var itemsList = await _itemDAO.GetAllItemsAsync();
            Items.Clear();
            ActiveItems.Clear();
            foreach (var item in itemsList.OrderBy(x => x.Id))
            {
                Items.Add(item);
            }

            // Sort active items alphabetically by name for the ComboBox
            var activeItemsList = itemsList
                .Where(x => !x.IsTombstone)
                .OrderBy(x => x.Content)
                .ToList();
            foreach (var item in activeItemsList)
            {
                ActiveItems.Add(item);
            }

            HasItems = itemsList.Count > 0;
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to load items: {ex.Message}");
            HasItems = false;
        }
    }

    private async System.Threading.Tasks.Task LoadOrdersAsync()
    {
        try
        {
            var ordersList = await _orderDAO.GetAllOrdersAsync();
            Orders.Clear();
            foreach (var order in ordersList.OrderBy(x => x.Id))
            {
                Orders.Add(order);
            }

            HasOrders = ordersList.Count > 0;
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to load orders: {ex.Message}");
            HasOrders = false;
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task PopulateInventoryAsync()
    {
        if (!AuthService.HasPermission("populate_inventory"))
        {
            ToastService.ShowWarning("You don't have permission to populate inventory");
            return;
        }

        try
        {
            var inventoryPath = "Data/inventory.json";
            if (!File.Exists(inventoryPath))
            {
                System.Console.WriteLine("[ERROR] inventory.json file not found");
                ToastService.ShowWarning("inventory.json file not found");
                return;
            }

            System.Console.WriteLine("[INFO] Loading inventory from JSON...");
            var jsonContent = await File.ReadAllTextAsync(inventoryPath);
            var inventoryItems = JsonSerializer.Deserialize<InventoryItem[]>(jsonContent);

            if (inventoryItems == null || inventoryItems.Length == 0)
            {
                System.Console.WriteLine("[WARNING] No items found in inventory.json");
                ToastService.ShowWarning("No items found in inventory file");
                return;
            }

            int addedCount = 0;
            foreach (var item in inventoryItems)
            {
                if (!string.IsNullOrEmpty(item.name))
                {
                    await _itemDAO.AddItemAsync(item.name, (decimal)item.price);
                    addedCount++;
                    System.Console.WriteLine($"[INFO] Added: {item.name} - ${item.price:F2}");
                }
            }

            await LoadItemsAsync();
            System.Console.WriteLine(
                $"[INFO] Successfully populated {addedCount} items from inventory"
            );
            ToastService.ShowSuccess($"Populated {addedCount} items from inventory");
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to populate inventory: {ex.Message}");
            ToastService.ShowWarning($"Failed to populate inventory: {ex.Message}");
        }
    }

    [RelayCommand]
    private void ToggleConsole()
    {
        IsConsoleVisible = !IsConsoleVisible;
        System.Console.WriteLine($"[INFO] Console {(IsConsoleVisible ? "shown" : "hidden")}");
    }

    [RelayCommand]
    private void ClearConsole()
    {
        ConsoleService.Clear();
        System.Console.WriteLine("[INFO] Console cleared");
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task CreateUserAsync()
    {
        if (!IsAdmin)
        {
            ToastService.ShowWarning("Only admins can create users");
            return;
        }

        if (string.IsNullOrWhiteSpace(NewUsername) || string.IsNullOrWhiteSpace(NewPassword))
        {
            ToastService.ShowWarning("Please enter both username and password");
            return;
        }

        try
        {
            var role = NewUserRoleIndex == 1 ? UserRole.Admin : UserRole.User;
            var userDAO = new UserDAO();

            await userDAO.AddUserAsync(NewUsername, NewPassword, role);
            ToastService.ShowSuccess($"User '{NewUsername}' created successfully");

            NewUsername = string.Empty;
            NewPassword = string.Empty;
            NewUserRoleIndex = 0;

            await LoadUsersAsync();
        }
        catch (Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to create user: {ex.Message}");
            ToastService.ShowWarning($"Failed to create user: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task DeleteUserAsync(ushort userId)
    {
        if (!IsAdmin)
        {
            ToastService.ShowWarning("Only admins can delete users");
            return;
        }

        try
        {
            var userDAO = new UserDAO();
            var users = await userDAO.GetAllUsersAsync();
            var userToDelete = users.FirstOrDefault(u => u.Id == userId);

            // Prevent deletion of admin users
            if (userToDelete?.Role == UserRole.Admin)
            {
                ToastService.ShowWarning("Cannot delete admin users");
                return;
            }

            await userDAO.DeleteUserAsync(userId);
            ToastService.ShowSuccess($"User deleted successfully");
            await LoadUsersAsync();
        }
        catch (Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to delete user: {ex.Message}");
            ToastService.ShowWarning($"Failed to delete user: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task ReadUsersAsync()
    {
        if (!IsAdmin)
        {
            ToastService.ShowWarning("Only admins can view users");
            return;
        }

        try
        {
            var userDAO = new UserDAO();
            var users = await userDAO.GetAllUsersAsync();

            System.Console.WriteLine("[INFO] Reading users from file...");

            if (users.Count == 0)
            {
                System.Console.WriteLine("[INFO] No users found");
                return;
            }

            var orderedUsers = users.OrderBy(x => x.Id).ToList();
            for (int i = 0; i < orderedUsers.Count; i++)
            {
                var user = orderedUsers[i];
                var status = user.IsTombstone ? "DELETED" : "ACTIVE";
                System.Console.WriteLine(
                    $"[ID: {user.Id}][{status}] - [Username: '{user.Username}'] - [Role: {user.Role}]"
                );
            }

            System.Console.WriteLine($"[INFO] Found {users.Count} users total");
        }
        catch (Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to read users: {ex.Message}");
        }
    }

    private async System.Threading.Tasks.Task LoadUsersAsync()
    {
        if (!IsAdmin)
            return;

        try
        {
            var userDAO = new UserDAO();
            var usersList = await userDAO.GetAllUsersAsync();
            Users.Clear();

            // Filter out admin users from the UI - they should not be visible or deletable
            foreach (var user in usersList.OrderBy(x => x.Id).Where(u => u.Role != UserRole.Admin))
            {
                Users.Add(user);
            }

            HasUsers = Users.Count > 0;
        }
        catch (Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to load users: {ex.Message}");
            HasUsers = false;
        }
    }

    [RelayCommand]
    private static void Quit()
    {
        System.Console.WriteLine("[INFO] Quitting application...");
        if (
            Application.Current?.ApplicationLifetime
            is IClassicDesktopStyleApplicationLifetime desktop
        )
        {
            desktop.Shutdown();
        }
    }

    private class InventoryItem
    {
        public string name { get; set; } = string.Empty;
        public double price { get; set; }
    }

    public class CartDisplayItem
    {
        public ushort ItemId { get; set; }
        public string ItemName { get; set; } = string.Empty;
    }
}

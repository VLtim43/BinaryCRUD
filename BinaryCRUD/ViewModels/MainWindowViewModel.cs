using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
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
    public ToastService ToastService { get; }

    public MainWindowViewModel(ItemDAO itemDAO)
    {
        _itemDAO = itemDAO;
        ToastService = new ToastService();
        _ = LoadItemsAsync();
    }

    [ObservableProperty]
    private string text = string.Empty;

    [ObservableProperty]
    private decimal priceInput = 0.0m;

    [ObservableProperty]
    private ObservableCollection<Item> items = new();

    [RelayCommand]
    private async System.Threading.Tasks.Task SaveAsync()
    {
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
            foreach (var item in itemsList.OrderBy(x => x.Id))
            {
                Items.Add(item);
            }
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to load items: {ex.Message}");
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
}

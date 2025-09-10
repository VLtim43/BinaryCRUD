using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using Avalonia;
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
            return;
        }

        try
        {
            System.Console.WriteLine(
                $"[INFO] Creating item: '{Text}' with price: ${PriceInput}"
            );
            await _itemDAO.AddItemAsync(Text, PriceInput);
            ToastService.ShowSuccess(
                $"Item '{Text}' (${PriceInput:F2}) created successfully"
            );
            Text = string.Empty;
            PriceInput = 0.0m;
            await LoadItemsAsync();
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine(
                $"[ERROR] Failed to create item: {ex.Message}"
            );
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
            System.Console.WriteLine($"[HEADER] Last Updated: {header.LastUpdated:yyyy-MM-dd HH:mm:ss} UTC");
            System.Console.WriteLine($"[HEADER] Header Size: 12 bytes");
            System.Console.WriteLine($"[HEADER] File Path: item.bin");

            ToastService.ShowSuccess($"Header: {header.Count} items, Updated: {header.LastUpdated:HH:mm:ss}");
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
            System.Console.WriteLine(
                $"[ERROR] Failed to read items: {ex.Message}"
            );
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task RefreshItemsAsync()
    {
        await LoadItemsAsync();
        ToastService.ShowSuccess("Items list refreshed");
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

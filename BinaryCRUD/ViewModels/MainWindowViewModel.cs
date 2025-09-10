using System;
using System.Collections.Generic;
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
    }

    [ObservableProperty]
    private string text = string.Empty;

    [ObservableProperty]
    private decimal priceInput = 0.0m;


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
                ToastService.ShowSuccess("No items found");
                return;
            }

            // Console output: Show all items
            for (int i = 0; i < items.Count; i++)
            {
                var item = items[i];
                var status = item.IsTombstone ? "DELETED" : "ACTIVE";
                System.Console.WriteLine(
                    $"[ID: {item.Id}][{status}] - [Name: '{item.Content}'] - [Price: ${item.Price:F2}]"
                );
            }

            System.Console.WriteLine($"[INFO] Found {items.Count} items total");

            // Toast: Show first 10 items
            var first10 = items.Take(10).ToList();
            var toastItems = string.Join(", ", first10.Select(item => 
                $"{item.Content}(${item.Price:F2}){(item.IsTombstone ? "❌" : "")}"
            ));
            
            var moreIndicator = items.Count > 10 ? $" +{items.Count - 10} more" : "";
            ToastService.ShowSuccess($"Items: {toastItems}{moreIndicator}");
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine(
                $"[ERROR] Failed to read items: {ex.Message}"
            );
            ToastService.ShowSuccess("Error reading items");
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

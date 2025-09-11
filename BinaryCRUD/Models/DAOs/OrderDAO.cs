using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class OrderDAO : FileBinaryDAO<Order>
{
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    public OrderDAO()
        : base("order.bin") { }

    public async Task AddOrderAsync(
        List<ushort> itemIds,
        float totalPrice,
        ushort userId,
        string? additionalInfo = null
    )
    {
        var order = new Order
        {
            ItemIds = itemIds,
            TotalPrice = totalPrice,
            UserId = userId,
            AdditionalInfo = additionalInfo,
        };
        await AddAsync(order);
    }

    // Convenience method for single item orders
    public async Task AddOrderAsync(
        ushort itemId,
        float totalPrice,
        ushort userId,
        string? additionalInfo = null
    )
    {
        var itemIds = new List<ushort> { itemId };
        await AddOrderAsync(itemIds, totalPrice, userId, additionalInfo);
    }

    // Convenience method for multiple instances of same item
    public async Task AddOrderAsync(
        ushort itemId,
        int quantity,
        float totalPrice,
        ushort userId,
        string? additionalInfo = null
    )
    {
        var itemIds = new List<ushort>();
        for (int i = 0; i < quantity; i++)
        {
            itemIds.Add(itemId);
        }
        await AddOrderAsync(itemIds, totalPrice, userId, additionalInfo);
    }

    public async Task<List<Order>> GetAllOrdersAsync()
    {
        return await GetAllAsync();
    }

    public async Task<FileHeader?> ReadHeaderAsync()
    {
        return await GetHeaderAsync();
    }

    public async Task DeleteOrderAsync(ushort orderId)
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(_filePath))
            {
                throw new InvalidOperationException("File does not exist");
            }

            var orders = await GetAllOrdersAsync();
            var orderToDelete = orders.FirstOrDefault(o => o.Id == orderId);

            if (orderToDelete == null)
            {
                throw new InvalidOperationException($"Order with ID {orderId} not found");
            }

            if (orderToDelete.IsTombstone)
            {
                throw new InvalidOperationException($"Order with ID {orderId} is already deleted");
            }

            orderToDelete.IsTombstone = true;

            await RewriteFileAsync(orders);
        }
        finally
        {
            _fileLock.Release();
        }
    }

    private async Task RewriteFileAsync(List<Order> orders)
    {
        var tempFilePath = _filePath + ".tmp";

        using (var stream = new FileStream(tempFilePath, FileMode.Create, FileAccess.Write))
        {
            var header = new FileHeader { Count = orders.Count };

            var headerBuffer = new byte[4];
            BitConverter.GetBytes(header.Count).CopyTo(headerBuffer, 0);
            await stream.WriteAsync(headerBuffer, 0, headerBuffer.Length);

            foreach (var order in orders)
            {
                var entityBytes = order.ToBytes();
                var lengthBytes = BitConverter.GetBytes(entityBytes.Length);
                await stream.WriteAsync(lengthBytes, 0, lengthBytes.Length);
                await stream.WriteAsync(entityBytes, 0, entityBytes.Length);
            }
        }

        File.Move(tempFilePath, _filePath, true);
    }

    public async Task DeleteFileAsync()
    {
        await _fileLock.WaitAsync();
        try
        {
            if (File.Exists(_filePath))
            {
                File.Delete(_filePath);
                System.Console.WriteLine($"[{GetType().Name}] File deleted: {_filePath}");
            }
            else
            {
                System.Console.WriteLine($"[{GetType().Name}] File does not exist: {_filePath}");
            }
        }
        finally
        {
            _fileLock.Release();
        }
    }
}

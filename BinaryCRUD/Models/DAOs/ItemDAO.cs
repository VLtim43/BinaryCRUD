using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class ItemDAO : FileBinaryDAO<Item>
{
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    public ItemDAO()
        : base("item.bin") { }

    public async Task AddItemAsync(string content, decimal price)
    {
        var item = new Item { Content = content, Price = price };
        await AddAsync(item);
    }

    public async Task<List<Item>> GetAllItemsAsync()
    {
        return await GetAllAsync();
    }

    public async Task<FileHeader?> ReadHeaderAsync()
    {
        return await GetHeaderAsync();
    }

    public async Task DeleteItemAsync(long itemId)
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(_filePath))
            {
                throw new InvalidOperationException("File does not exist");
            }

            var items = await GetAllItemsAsync();
            var itemToDelete = items.FirstOrDefault(i => i.Id == itemId);
            
            if (itemToDelete == null)
            {
                throw new InvalidOperationException($"Item with ID {itemId} not found");
            }

            if (itemToDelete.IsTombstone)
            {
                throw new InvalidOperationException($"Item with ID {itemId} is already deleted");
            }

            itemToDelete.IsTombstone = true;

            await RewriteFileAsync(items);
        }
        finally
        {
            _fileLock.Release();
        }
    }

    private async Task RewriteFileAsync(List<Item> items)
    {
        var tempFilePath = _filePath + ".tmp";
        
        using (var stream = new FileStream(tempFilePath, FileMode.Create, FileAccess.Write))
        {
            var header = new FileHeader
            {
                Count = items.Count,
                LastUpdated = DateTime.UtcNow
            };

            var headerBuffer = new byte[12];
            BitConverter.GetBytes(header.Count).CopyTo(headerBuffer, 0);
            BitConverter.GetBytes(header.LastUpdated.Ticks).CopyTo(headerBuffer, 4);
            await stream.WriteAsync(headerBuffer, 0, headerBuffer.Length);

            foreach (var item in items)
            {
                var entityBytes = item.ToBytes();
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
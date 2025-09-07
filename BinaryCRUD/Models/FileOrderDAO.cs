using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class FileOrderDAO : IOrderDAO, IDisposable
{
    private const string DataDirectory = "Data";
    private const string OrdersFileName = "orders.bin";
    private readonly string _filePath;
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    public FileOrderDAO()
    {
        if (!Directory.Exists(DataDirectory))
            Directory.CreateDirectory(DataDirectory);

        _filePath = Path.Combine(DataDirectory, OrdersFileName);
    }

    public async Task AddOrderAsync(string content)
    {
        await _fileLock.WaitAsync();
        try
        {
            var header = await ReadOrCreateHeaderAsync();
            var previousCount = header.Count;
            header.Count++;
            header.LastUpdated = DateTime.UtcNow;

            Console.WriteLine($"[DAO] Updating header: {previousCount} -> {header.Count} orders");

            using var stream = new FileStream(_filePath, FileMode.Create, FileAccess.Write);

            // Write header
            await WriteHeaderAsync(stream, header);

            // Copy existing orders if file existed
            if (File.Exists(_filePath + ".tmp"))
            {
                using var tempStream = new FileStream(
                    _filePath + ".tmp",
                    FileMode.Open,
                    FileAccess.Read
                );
                tempStream.Seek(GetHeaderSize(), SeekOrigin.Begin);
                await tempStream.CopyToAsync(stream);
            }

            // Append new order
            await WriteOrderAsync(stream, content);
            Console.WriteLine($"[DAO] Order saved to: {_filePath}");
        }
        finally
        {
            _fileLock.Release();
        }
    }

    private async Task<OrderHeader> ReadOrCreateHeaderAsync()
    {
        if (!File.Exists(_filePath))
            return new OrderHeader { Count = 0 };

        // Create temp copy for reading existing data
        File.Copy(_filePath, _filePath + ".tmp", true);

        using var stream = new FileStream(_filePath + ".tmp", FileMode.Open, FileAccess.Read);
        return await ReadHeaderAsync(stream);
    }

    private async Task<OrderHeader> ReadHeaderAsync(Stream stream)
    {
        var buffer = new byte[GetHeaderSize()];
        await stream.ReadExactlyAsync(buffer);

        var count = BitConverter.ToInt32(buffer, 0);
        var ticks = BitConverter.ToInt64(buffer, 4);

        return new OrderHeader
        {
            Count = count,
            LastUpdated = new DateTime(ticks, DateTimeKind.Utc),
        };
    }

    private async Task WriteHeaderAsync(Stream stream, OrderHeader header)
    {
        var buffer = new byte[GetHeaderSize()];
        BitConverter.GetBytes(header.Count).CopyTo(buffer, 0);
        BitConverter.GetBytes(header.LastUpdated.Ticks).CopyTo(buffer, 4);

        await stream.WriteAsync(buffer, 0, buffer.Length);
    }

    private async Task WriteOrderAsync(Stream stream, string content)
    {
        var contentBytes = Encoding.UTF8.GetBytes(content);
        var lengthBytes = BitConverter.GetBytes(contentBytes.Length);

        await stream.WriteAsync(lengthBytes, 0, lengthBytes.Length);
        await stream.WriteAsync(contentBytes, 0, contentBytes.Length);
    }

    private static int GetHeaderSize() => sizeof(int) + sizeof(long); // Count + LastUpdated ticks

    protected virtual void Dispose(bool disposing)
    {
        if (disposing)
        {
            _fileLock?.Dispose();
            if (File.Exists(_filePath + ".tmp"))
                File.Delete(_filePath + ".tmp");
        }
    }

    public void Dispose()
    {
        Dispose(true);
        GC.SuppressFinalize(this);
    }
}

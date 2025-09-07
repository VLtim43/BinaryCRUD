using System;
using System.Collections.Generic;
using System.IO;
using System.Threading;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public abstract class FileBinaryDAO<T> : IFileDAO<T>, IDisposable
    where T : ISerializable
{
    private const string DataDirectory = "Data";
    protected readonly string _filePath;
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    protected FileBinaryDAO(string fileName)
    {
        if (!Directory.Exists(DataDirectory))
            Directory.CreateDirectory(DataDirectory);

        _filePath = Path.Combine(DataDirectory, fileName);
    }

    public async Task AddAsync(T entity)
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(_filePath))
            {
                await CreateNewFileWithHeaderAsync();
            }

            // Read current header
            var header = await ReadHeaderAsync();
            var previousCount = header.Count;
            header.Count++;
            header.LastUpdated = DateTime.UtcNow;

            Console.WriteLine(
                $"[{GetType().Name}] Updating header: {previousCount} -> {header.Count} entities"
            );

            await UpdateHeaderAsync(header);
            await AppendEntityToFileAsync(entity);

            Console.WriteLine($"[{GetType().Name}] Entity appended to: {_filePath}");
        }
        finally
        {
            _fileLock.Release();
        }
    }

    private async Task CreateNewFileWithHeaderAsync()
    {
        var header = new FileHeader { Count = 0 };
        using var stream = new FileStream(_filePath, FileMode.Create, FileAccess.Write);
        await WriteHeaderAsync(stream, header);
        Console.WriteLine($"[{GetType().Name}] Created new file with header: {_filePath}");
    }

    private async Task<FileHeader> ReadHeaderAsync()
    {
        using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.Read);
        return await ReadHeaderFromStreamAsync(stream);
    }

    private async Task UpdateHeaderAsync(FileHeader header)
    {
        using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.ReadWrite);
        stream.Seek(0, SeekOrigin.Begin);
        await WriteHeaderAsync(stream, header);
    }

    private async Task AppendEntityToFileAsync(T entity)
    {
        using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.Write);
        stream.Seek(0, SeekOrigin.End);
        await WriteEntityAsync(stream, entity);
    }

    private async Task<FileHeader> ReadHeaderFromStreamAsync(Stream stream)
    {
        var buffer = new byte[GetHeaderSize()];
        await stream.ReadExactlyAsync(buffer);

        var count = BitConverter.ToInt32(buffer, 0);
        var ticks = BitConverter.ToInt64(buffer, 4);

        return new FileHeader
        {
            Count = count,
            LastUpdated = new DateTime(ticks, DateTimeKind.Utc),
        };
    }

    private async Task WriteHeaderAsync(Stream stream, FileHeader header)
    {
        var buffer = new byte[GetHeaderSize()];
        BitConverter.GetBytes(header.Count).CopyTo(buffer, 0);
        BitConverter.GetBytes(header.LastUpdated.Ticks).CopyTo(buffer, 4);

        await stream.WriteAsync(buffer, 0, buffer.Length);
    }

    private async Task WriteEntityAsync(Stream stream, T entity)
    {
        var entityBytes = entity.ToBytes();
        var lengthBytes = BitConverter.GetBytes(entityBytes.Length);

        await stream.WriteAsync(lengthBytes, 0, lengthBytes.Length);
        await stream.WriteAsync(entityBytes, 0, entityBytes.Length);
    }

    public async Task<List<T>> GetAllAsync()
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(_filePath))
            {
                Console.WriteLine($"[{GetType().Name}] File does not exist: {_filePath}");
                return new List<T>();
            }

            using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.Read);

            // Read header
            var header = await ReadHeaderFromStreamAsync(stream);
            Console.WriteLine(
                $"[{GetType().Name}] Reading {header.Count} entities from: {_filePath}"
            );

            var entities = new List<T>();

            // Read each entity
            for (int i = 0; i < header.Count; i++)
            {
                var entity = await ReadEntityFromStreamAsync(stream);
                entities.Add(entity);
            }

            return entities;
        }
        finally
        {
            _fileLock.Release();
        }
    }

    private async Task<T> ReadEntityFromStreamAsync(Stream stream)
    {
        // Read entity length
        var lengthBuffer = new byte[sizeof(int)];
        await stream.ReadExactlyAsync(lengthBuffer);
        var entityLength = BitConverter.ToInt32(lengthBuffer, 0);

        // Read entity data
        var entityBuffer = new byte[entityLength];
        await stream.ReadExactlyAsync(entityBuffer);

        // Deserialize entity
        var entity = (T)Activator.CreateInstance(typeof(T))!;
        entity.FromBytes(entityBuffer);
        return entity;
    }

    private static int GetHeaderSize() => sizeof(int) + sizeof(long);

    protected virtual void Dispose(bool disposing)
    {
        if (disposing)
        {
            _fileLock?.Dispose();
        }
    }

    public void Dispose()
    {
        Dispose(true);
        GC.SuppressFinalize(this);
    }
}

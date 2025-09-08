using System;

namespace BinaryCRUD.Models;

public class FileHeader
{
    public int Count { get; set; }
    public DateTime LastUpdated { get; set; } = DateTime.UtcNow;
}

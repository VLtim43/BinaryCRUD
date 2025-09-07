using System;

namespace BinaryCRUD.Models;

public class OrderHeader
{
    public int Count { get; set; }
    public DateTime LastUpdated { get; set; } = DateTime.UtcNow;
}

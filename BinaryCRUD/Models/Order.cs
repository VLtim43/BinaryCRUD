using System;

namespace BinaryCRUD.Models;

public class Order
{
    public string Content { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
}

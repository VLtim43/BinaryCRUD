using System.Collections.Generic;

namespace BinaryCRUD.Models;

public interface IOrderDAO
{
    System.Threading.Tasks.Task AddOrderAsync(string content);
    System.Threading.Tasks.Task<List<Order>> GetAllOrdersAsync();
}

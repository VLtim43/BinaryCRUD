using System.Collections.Generic;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class OrderDAO : FileBinaryDAO<Order>
{
    public OrderDAO()
        : base("orders.bin") { }

    public async Task AddOrderAsync(string content)
    {
        var order = new Order { Content = content };
        await AddAsync(order);
    }

    public async Task<List<Order>> GetAllOrdersAsync()
    {
        return await GetAllAsync();
    }
}

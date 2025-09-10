using System.Collections.Generic;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class ItemDAO : FileBinaryDAO<Item>
{
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
}
namespace BinaryCRUD.Models;

public interface IOrderDAO
{
    System.Threading.Tasks.Task AddOrderAsync(string content);
}

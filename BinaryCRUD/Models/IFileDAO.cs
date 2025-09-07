namespace BinaryCRUD.Models;

public interface IFileDAO<T> where T : ISerializable
{
    System.Threading.Tasks.Task AddAsync(T entity);
}
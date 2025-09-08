using System.Collections.Generic;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public interface InterfaceFileDAO<T>
    where T : InterfaceSerializable
{
    Task AddAsync(T entity);
    Task<List<T>> GetAllAsync();
}

using System.Collections.Generic;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class UserDAO : FileBinaryDAO<User>
{
    public UserDAO()
        : base("users.bin") { }

    public async Task AddUserAsync(string content)
    {
        var user = new User { Content = content };
        await AddAsync(user);
    }

    public async Task<List<User>> GetAllUsersAsync()
    {
        return await GetAllAsync();
    }
}
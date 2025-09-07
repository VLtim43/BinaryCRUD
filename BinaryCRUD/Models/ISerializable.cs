namespace BinaryCRUD.Models;

public interface ISerializable
{
    byte[] ToBytes();
    void FromBytes(byte[] data);
}
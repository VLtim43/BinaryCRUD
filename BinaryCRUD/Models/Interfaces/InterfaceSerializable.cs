namespace BinaryCRUD.Models;

public interface InterfaceSerializable
{
    byte[] ToBytes();
    void FromBytes(byte[] data);
}

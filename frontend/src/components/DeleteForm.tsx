import { h } from "preact";
import { Button } from "./Button";
import { Input } from "./Input";
import { createIdInputHandler } from "../utils/formatters";

interface DeleteFormProps {
  deleteId: string;
  setDeleteId: (value: string) => void;
  onDelete: () => void;
  entityName?: string;
}

export const DeleteForm = ({
  deleteId,
  setDeleteId,
  onDelete,
  entityName = "Record",
}: DeleteFormProps) => {
  return (
    <div className="input-box">
      <Input
        id="delete-record-id"
        placeholder="Enter Record ID"
        value={deleteId}
        onChange={createIdInputHandler(setDeleteId)}
      />
      <Button variant="danger" onClick={onDelete}>
        Delete {entityName}
      </Button>
    </div>
  );
};

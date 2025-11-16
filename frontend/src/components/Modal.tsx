import { h } from "preact";
import { Button } from "./Button";

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: any;
}

export const Modal = ({ isOpen, onClose, title, children }: ModalProps) => {
  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{title}</h3>
          <Button onClick={onClose} size="small" className="modal-close">
            ×
          </Button>
        </div>
        <div className="modal-body">{children}</div>
      </div>
    </div>
  );
};

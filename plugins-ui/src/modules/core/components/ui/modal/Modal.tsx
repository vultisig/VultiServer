import { ReactNode } from "react";
import "./Modal.css";
import CloseIcon from "@/assets/Close.svg?react";
import Button from "../button/Button";

type ModalProps = {
  isOpen: boolean;
  children: ReactNode;
  onClose: () => void;
};

function Modal({ isOpen, children, onClose }: ModalProps) {
  if (!isOpen) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <Button
          type="button"
          styleType="tertiary"
          size="medium"
          className="modal-close"
          onClick={onClose}
        >
          <CloseIcon />
        </Button>
        {children}
      </div>
    </div>
  );
}

export default Modal;

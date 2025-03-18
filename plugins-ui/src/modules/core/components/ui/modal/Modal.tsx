import { ReactNode } from "react";
import "./Modal.css";
import CloseIcon from "@/assets/Close.svg?react";
import Button from "../button/Button";

type ModalProps = {
  isOpen: boolean;
  children: ReactNode;
  onClose: () => void;
  variant?: "modal" | "panel";
};

function Modal({ isOpen, children, onClose, variant = "modal" }: ModalProps) {
  if (!isOpen) return null;

  return (
    <div
      className={`modal-overlay ${variant === "panel" ? "panel-overlay" : ""}`}
      role="dialog"
    >
      <div
        className={`modal-content ${variant === "panel" ? "panel-content" : ""}`}
        onClick={(e) => e.stopPropagation()} // Prevent closing when clicking inside
      >
        <Button
          ariaLabel="Close modal"
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

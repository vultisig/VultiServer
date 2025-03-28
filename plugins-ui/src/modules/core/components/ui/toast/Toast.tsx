import { useEffect } from "react";
import "./Toast.css";
import Button from "../button/Button";
import CloseIcon from "@/assets/Close.svg?react";

interface ToastProps {
  title: string;
  message?: string;
  type: "success" | "warning" | "error";
  onClose: () => void;
  duration?: number; // Auto-close time (ms)
}

const Toast = ({
  title,
  message,
  type,
  onClose,
  duration = 5000,
}: ToastProps) => {
  useEffect(() => {
    const timer = setTimeout(onClose, duration);
    return () => clearTimeout(timer); // Cleanup on unmount
  }, [onClose, duration]);

  return (
    <div className={`toast toast-${type}`}>
      <div className="message">{title}</div>
      <div>{message}</div>
      <Button
        ariaLabel="Close message"
        style={{ paddingRight: 0 }}
        size="small"
        type="button"
        styleType="tertiary"
        onClick={onClose}
      >
        <CloseIcon />
      </Button>
    </div>
  );
};

export default Toast;

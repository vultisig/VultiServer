import { ReactNode } from "react";
import "./Modal.css";
import closeIcon from "@/assets/Close.svg";
import Button from "../button/Button";

type ModalProps = {
    isOpen: boolean, children: ReactNode,
    onClose: () => void;
}

function Modal({ isOpen, children, onClose }: ModalProps) {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <Button type="tertiary" size="medium" className="modal-close" onClick={onClose}>
                    <img src={closeIcon} alt="" />
                </Button>
                {children}
            </div>
        </div>
    );
}

export default Modal;

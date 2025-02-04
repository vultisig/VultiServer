import { ReactNode, useState } from "react";
import SelectorModal from "./SelectorModal";
import Button from "../button/Button";
import "./SelectorTriggerButton.css";
import ChevronDown from "@/assets/ChevronDown.svg?react";
import { Option } from "../../../models/select-modal";

type SelectorTriggerButtonProps = {
  modalHeader: string;
  placeholder: string;
  data: Option[];
  value: string;
  onChange: (value: string) => void;
  children: ReactNode;
};

const SelectorTriggerButton = ({
  modalHeader,
  placeholder,
  data,
  value,
  onChange,
  children,
}: SelectorTriggerButtonProps) => {
  const [isModalOpen, setModalOpen] = useState(false);

  return (
    <>
      <Button
        type="button"
        styleType="tertiary"
        size="small"
        style={{ backgroundColor: "#11284A" }}
        onClick={() => setModalOpen(true)}
      >
        {value && <>{children}</>}
        <ChevronDown width="20px" height="20px" />
      </Button>

      <SelectorModal
        modalHeader={modalHeader}
        placeholder={placeholder}
        isOpen={isModalOpen}
        options={data}
        onClose={() => setModalOpen(false)}
        onSelect={onChange}
      />
    </>
  );
};

export default SelectorTriggerButton;

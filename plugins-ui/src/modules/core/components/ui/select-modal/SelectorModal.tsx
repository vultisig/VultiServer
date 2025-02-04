import React, { useState } from "react";
import "./SelectorModal.css";
import Button from "../button/Button";
import CloseIcon from "@/assets/Close.svg?react";
import { Option } from "@/modules/core/models/select-modal";
import { Input } from "../input/Input";

type SelectorModalProps = {
  modalHeader: string;
  placeholder: string;
  isOpen: boolean;
  options: Option[];
  onClose: () => void;
  onSelect: (selectedOption: string) => void;
};

const SelectorModal: React.FC<SelectorModalProps> = ({
  modalHeader,
  placeholder,
  isOpen,
  options,
  onClose,
  onSelect,
}) => {
  const [search, setSearch] = useState("");

  // Filter options based on the search input
  const filteredOptions = options.filter((option) =>
    option.name.toLowerCase().includes(search.toLowerCase())
  );

  if (!isOpen) return null;

  return (
    <div className="modal-backdrop">
      <div className="modal">
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          className="modal-close"
          onClick={onClose}
        >
          <CloseIcon />
        </Button>
        <div className="modal-header">
          <h2>{modalHeader}</h2>
          <Input
            id="seatch"
            name="search"
            label=""
            type="text"
            validation={{
              value: search,
              onChange: (e) => setSearch(e.target.value),
            }}
            placeholder={placeholder}
          />
        </div>
        <ul className="modal-options">
          {filteredOptions.length > 0 ? (
            filteredOptions.map((option) => (
              <li
                tabIndex={0}
                key={option.id}
                className="modal-option"
                onClick={() => {
                  onSelect(option.id);
                  onClose();
                }}
              >
                {option.image}&nbsp;
                {option.name}
              </li>
            ))
          ) : (
            <li className="modal-no-options">No matching options</li>
          )}
        </ul>
      </div>
    </div>
  );
};

export default SelectorModal;

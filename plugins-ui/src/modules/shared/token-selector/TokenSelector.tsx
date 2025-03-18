import Button from "@/modules/core/components/ui/button/Button";
import ChevronDown from "@/assets/ChevronDown.svg?react";
import Search from "@/assets/Search.svg?react";
import { cloneElement, useState } from "react";
import { supportedTokens } from "@/modules/shared/data/tokens";
import Modal from "@/modules/core/components/ui/modal/Modal";

type TokenSelectorProps = {
  value: string;
  onChange: (tokenId: string) => void;
};

const TokenSelector = ({ value, onChange }: TokenSelectorProps) => {
  const [isModalOpen, setModalOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [token, setToken] = useState(supportedTokens[value]);

  const filteredOptions = Object.values(supportedTokens).filter((option) =>
    option.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleSelect = (optionId: string) => {
    onChange(optionId);
    setToken(supportedTokens[optionId]);
    setModalOpen(false);
  };

  return (
    <>
      <Button
        ariaLabel="Open modal"
        type="button"
        styleType="tertiary"
        size="small"
        style={{
          justifyContent: "space-between",
          backgroundColor: "#061B3A",
          border: "1px solid #11284A",
          borderRadius: "12px",
          width: "100%",
        }}
        onClick={() => setModalOpen(true)}
      >
        {token && cloneElement(token.image, { width: 24, height: 24 })}&nbsp;
        {token?.name || "Unknown token"}
        <ChevronDown width="20px" height="20px" />
      </Button>
      <Modal
        isOpen={isModalOpen}
        onClose={() => setModalOpen(false)}
        variant="modal"
      >
        <div className="modal-header">
          <h2>Select a token</h2>
          <div className="input-container">
            <input
              id="seatch"
              name="search"
              type="text"
              placeholder="Search by token"
              onChange={(e) => setSearch(e.target.value)}
            />
            <Search className="icon" width="20px" height="20px" />
          </div>
        </div>
        <ul className="modal-options">
          {filteredOptions.length > 0 ? (
            filteredOptions.map((option) => (
              <li
                tabIndex={0}
                key={option.id}
                className="modal-option"
                onClick={() => handleSelect(option.id)}
              >
                {option.image}&nbsp;
                {option.name}
              </li>
            ))
          ) : (
            <li className="modal-no-options">No matching options</li>
          )}
        </ul>
      </Modal>
    </>
  );
};

export default TokenSelector;

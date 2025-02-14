import { useState, ReactNode } from "react";
import "./Accordion.css";
import ChevronDown from "@/assets/ChevronDown.svg?react";
import ChevronRight from "@/assets/ChevronRight.svg?react";
import Button from "../button/Button";

type AccordionProps = {
  header: ReactNode;
  expandButton: {
    text: string;
    style?: {};
  };
  children: ReactNode;
};

const Accordion = ({ header, expandButton, children }: AccordionProps) => {
  const [isOpen, setIsOpen] = useState(false);

  const toggleAccordion = () => {
    setIsOpen(!isOpen);
  };

  return (
    <div className="accordion">
      <div className="accordion-header">
        {header}
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={expandButton.style}
          onClick={toggleAccordion}
        >
          {expandButton.text}
          {isOpen && <ChevronDown width="16px" height="16px" />}
          {!isOpen && <ChevronRight width="16px" height="16px" />}
        </Button>
      </div>
      {isOpen && <div className="accordion-content">{children}</div>}
    </div>
  );
};

export default Accordion;

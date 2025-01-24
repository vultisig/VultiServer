import { useState, ReactNode } from "react";
import "./Accordion.css";
import chevronDown from "@/assets/ChevronDown.svg";
import chevronRight from "@/assets/ChevronRight.svg";
import Button from "../button/Button";

type AccordionProps = {
    header: ReactNode,
    expandButton: {
        text: string,
        style?: {}
    }
    children: ReactNode
}

const Accordion = ({ header, expandButton, children }: AccordionProps) => {
    const [isOpen, setIsOpen] = useState(false);

    const toggleAccordion = () => {
        setIsOpen(!isOpen);
    };

    return (
        <div className={`accordion ${isOpen ? "expanded" : ""}`}>
            <div className="accordion-header">
                {header}
                <Button type="tertiary" size='small' style={expandButton.style} onClick={toggleAccordion}>
                    {expandButton.text}
                    <img src={isOpen ? chevronDown : chevronRight} alt="" width="16px" height="16px" />
                </Button>
            </div>
            {isOpen && <div className="accordion-content">{children}</div>}
        </div>
    );
};

export default Accordion;

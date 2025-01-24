import { ReactNode } from "react";
import "./Button.css";

type ButtonProps = {
    type: "primary" | "secondary" | "tertiary",
    size: "small" | "medium"
    children: ReactNode,
    className?: string,
    onClick: () => any,
    style?: {},
}

const Button = ({ type, size, children, className, onClick, style }: ButtonProps) => {

    return (
        <button
            onClick={onClick}
            className={`button ${type} ${size} ${className}`}
            style={style}
        >
            {children}
        </button>
    );
};

export default Button;
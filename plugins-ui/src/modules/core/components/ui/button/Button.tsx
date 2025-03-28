import { ReactNode } from "react";
import "./Button.css";

type ButtonProps = {
  type: "button" | "submit";
  styleType: "primary" | "secondary" | "tertiary";
  size: "mini" | "small" | "medium";
  children: ReactNode;
  className?: string;
  style?: {};
  ariaLabel?: string;
  onClick?: () => any;
};

const Button = ({
  type,
  styleType,
  size,
  children,
  className,
  style,
  ariaLabel,
  onClick,
}: ButtonProps) => {
  return (
    <button
      type={type}
      onClick={onClick}
      className={`button ${styleType} ${size} ${className}`}
      style={style}
      aria-label={ariaLabel}
    >
      {children}
    </button>
  );
};

export default Button;

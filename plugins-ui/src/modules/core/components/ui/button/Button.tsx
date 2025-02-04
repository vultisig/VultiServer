import { ReactNode } from "react";
import "./Button.css";

type ButtonProps = {
  type: "button" | "submit";
  styleType: "primary" | "secondary" | "tertiary";
  size: "small" | "medium";
  children: ReactNode;
  className?: string;
  style?: {};
  onClick?: () => any;
};

const Button = ({
  type,
  styleType,
  size,
  children,
  className,
  onClick,
  style,
}: ButtonProps) => {
  return (
    <button
      type={type}
      onClick={onClick}
      className={`button ${styleType} ${size} ${className}`}
      style={style}
    >
      {children}
    </button>
  );
};

export default Button;

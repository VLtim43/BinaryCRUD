import { h } from "preact";

interface ButtonProps {
  onClick?: () => void;
  disabled?: boolean;
  variant?: "primary" | "danger" | "default";
  size?: "default" | "small";
  className?: string;
  style?: any;
  children: any;
}

export const Button = ({
  onClick,
  disabled = false,
  variant = "default",
  size = "default",
  className = "",
  style,
  children,
}: ButtonProps) => {
  const variantClass = variant === "danger" ? "btn-danger" : variant === "primary" ? "btn-primary" : "";
  const sizeClass = size === "small" ? "btn-small" : "";
  const classes = `btn ${variantClass} ${sizeClass} ${className}`.trim();

  return (
    <button className={classes} onClick={onClick} disabled={disabled} style={style}>
      {children}
    </button>
  );
};

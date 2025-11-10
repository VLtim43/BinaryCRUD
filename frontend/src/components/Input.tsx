import { h } from "preact";

interface InputProps {
  value: string;
  onChange: (e: any) => void;
  placeholder?: string;
  type?: string;
  className?: string;
  style?: any;
  id?: string;
  name?: string;
  autoComplete?: string;
}

export const Input = ({
  value,
  onChange,
  placeholder = "",
  type = "text",
  className = "",
  style,
  id,
  name,
  autoComplete = "off",
}: InputProps) => {
  const classes = `input ${className}`.trim();

  return (
    <input
      id={id}
      name={name}
      className={classes}
      type={type}
      placeholder={placeholder}
      value={value}
      onChange={onChange}
      autoComplete={autoComplete}
      style={style}
    />
  );
};

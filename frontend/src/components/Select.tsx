import { h } from "preact";

interface SelectProps {
  value: string;
  onChange: (e: any) => void;
  options: Array<{ value: string | number; label: string }>;
  placeholder?: string;
  className?: string;
  style?: any;
}

export const Select = ({
  value,
  onChange,
  options,
  placeholder = "Select...",
  className = "",
  style,
}: SelectProps) => {
  const classes = `cart-select ${className}`.trim();

  return (
    <select className={classes} value={value} onChange={onChange} style={style}>
      <option value="">{placeholder}</option>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
};

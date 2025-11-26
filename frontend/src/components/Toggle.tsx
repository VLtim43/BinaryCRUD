import { h } from "preact";

interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  disabled?: boolean;
}

export const Toggle = ({ checked, onChange, label, disabled }: ToggleProps) => {
  return (
    <label className="toggle-container">
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange((e.target as HTMLInputElement).checked)}
        disabled={disabled}
        className="toggle-input"
      />
      <span className="toggle-slider" />
      {label && <span className="toggle-label">{label}</span>}
    </label>
  );
};

import { h } from "preact";

interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  disabled?: boolean;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

export const Toggle = ({ checked, onChange, label, disabled, onMouseEnter, onMouseLeave }: ToggleProps) => {
  return (
    <label className="toggle-container" onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
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

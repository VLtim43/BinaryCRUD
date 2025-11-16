import { h } from "preact";
import "./Badge.scss";

interface BadgeProps {
  variant?: "item" | "promo" | "default";
  children: preact.ComponentChildren;
}

export const Badge = ({ variant = "default", children }: BadgeProps) => {
  return <span className={`badge badge-${variant}`}>{children}</span>;
};

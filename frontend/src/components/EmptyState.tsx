import { h } from "preact";
import "./EmptyState.scss";

interface EmptyStateProps {
  message: string;
}

export const EmptyState = ({ message }: EmptyStateProps) => {
  return <div className="empty-state">{message}</div>;
};

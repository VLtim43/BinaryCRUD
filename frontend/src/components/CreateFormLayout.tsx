import { h } from "preact";
import "./CreateFormLayout.scss";

interface CreateFormLayoutProps {
  title: string;
  submitDisabled: boolean;
  onSubmit: () => void;
  headerInputs: preact.ComponentChildren;
  totalLabel: string;
  totalAmount: string;
  contentLabel: string;
  contentEmpty: boolean;
  emptyMessage: string;
  children: preact.ComponentChildren;
  footer: preact.ComponentChildren;
}

export const CreateFormLayout = ({
  title,
  submitDisabled,
  onSubmit,
  headerInputs,
  totalLabel,
  totalAmount,
  contentLabel,
  contentEmpty,
  emptyMessage,
  children,
  footer,
}: CreateFormLayoutProps) => {
  return (
    <div className="create-form-container">
      <div className="create-form">
        {/* Fixed Header */}
        <div className="create-form-header">
          <div className="create-form-title-row">
            <h3>{title}</h3>
            <button
              className="btn btn-primary"
              onClick={onSubmit}
              disabled={submitDisabled}
            >
              Submit
            </button>
          </div>
          {headerInputs}
          <div className="cart-total">{totalLabel}: ${totalAmount}</div>
        </div>

        {/* Scrollable Content */}
        <div className="create-form-content">
          <h4 className="create-form-content-label">{contentLabel}</h4>
          <div className="create-form-items-container">
            {contentEmpty ? (
              <div className="empty-state">{emptyMessage}</div>
            ) : (
              <div className="create-form-items-list">{children}</div>
            )}
          </div>
        </div>

        {/* Fixed Footer */}
        <div className="create-form-footer">{footer}</div>
      </div>
    </div>
  );
};

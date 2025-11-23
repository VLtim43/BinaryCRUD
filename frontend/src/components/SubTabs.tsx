import { h } from "preact";
import { Button } from "./Button";

export interface SubTab {
  id: string;
  label: string;
}

interface SubTabsProps {
  tabs: SubTab[];
  activeTab: string;
  onTabChange: (tabId: string) => void;
}

export const SubTabs = ({ tabs, activeTab, onTabChange }: SubTabsProps) => {
  return (
    <div className="sub_tabs">
      {tabs.map((tab) => (
        <Button
          key={tab.id}
          className={`tab ${activeTab === tab.id ? "active" : ""}`}
          onClick={() => onTabChange(tab.id)}
        >
          {tab.label}
        </Button>
      ))}
    </div>
  );
};

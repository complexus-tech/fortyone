import type { ReactNode } from "react";
import type { ViewProps } from "react-native";
import type { TabsControlProps } from "./tabs-control.types";
import { createContext, useContext, useState } from "react";
import { View } from "react-native";
import { Row } from "./row";
import { TabsControl } from "./tabs-control";

type TabsContextValue = {
  activeTab: string;
  onTabChange: (value: string) => void;
};

const TabsContext = createContext<TabsContextValue | undefined>(undefined);

const useTabsContext = () => {
  const context = useContext(TabsContext);
  if (!context) {
    throw new Error("Tabs compound components must be used within Tabs");
  }
  return context;
};

type TabsProps = {
  children: ReactNode;
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
};

export const Tabs = ({
  children,
  value,
  defaultValue,
  onValueChange,
}: TabsProps) => {
  const [internalValue, setInternalValue] = useState(defaultValue || "");

  const activeTab = value !== undefined ? value : internalValue;

  const handleTabChange = (newValue: string) => {
    if (value === undefined) {
      setInternalValue(newValue);
    }
    onValueChange?.(newValue);
  };

  return (
    <TabsContext.Provider value={{ activeTab, onTabChange: handleTabChange }}>
      {children}
    </TabsContext.Provider>
  );
};

type TabsListProps = Omit<ViewProps, "children"> &
  Pick<TabsControlProps, "options" | "labelSize">;

const TabsList = ({
  options,
  labelSize,
  accessibilityLabel,
  ...props
}: TabsListProps) => {
  const { activeTab, onTabChange } = useTabsContext();

  return (
    <Row align="center" asContainer className="mb-[4px]" {...props}>
      <TabsControl
        options={options}
        value={activeTab}
        onValueChange={onTabChange}
        labelSize={labelSize}
        accessibilityLabel={accessibilityLabel}
      />
    </Row>
  );
};

type TabPanelProps = {
  children: ReactNode;
  value: string;
};

const TabPanel = ({ children, value }: TabPanelProps) => {
  const { activeTab } = useTabsContext();
  if (activeTab !== value) {
    return null;
  }

  return <View style={{ flex: 1 }}>{children}</View>;
};

Tabs.List = TabsList;
Tabs.Panel = TabPanel;

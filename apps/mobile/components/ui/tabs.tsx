import React, { createContext, useContext, useState, ReactNode } from "react";
import { View, Pressable, ViewProps } from "react-native";
import { Text } from "./text";
import { Row } from "./row";
import { cn } from "@/lib/utils";

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

type TabsListProps = ViewProps & {
  children: ReactNode;
};

const TabsList = ({ children, ...props }: TabsListProps) => {
  return (
    <Row gap={1} asContainer className="mb-1" {...props}>
      {children}
    </Row>
  );
};

type TabProps = {
  children: ReactNode;
  value: string;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  className?: string;
};

const Tab = ({ children, value, leftIcon, rightIcon, className }: TabProps) => {
  const { activeTab, onTabChange } = useTabsContext();
  const isActive = activeTab === value;

  return (
    <Pressable
      onPress={() => onTabChange(value)}
      className={cn(
        "active:bg-gray-50 px-4 dark:active:bg-dark py-2.5 rounded-full flex-row justify-center gap-2",
        {
          "bg-gray-50 dark:bg-dark-200/80 border dark:border-dark-50/70 border-gray-200/60":
            isActive,
        },
        className
      )}
    >
      {leftIcon}
      <Text fontSize="sm" color={isActive ? undefined : "muted"}>
        {children}
      </Text>
      {rightIcon}
    </Pressable>
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
Tabs.Tab = Tab;
Tabs.Panel = TabPanel;

import React, { createContext, useContext, useState, ReactNode } from "react";
import { View, Pressable, ViewProps } from "react-native";
import { Text } from "./Text";
import { Row } from "./row";
import { cn } from "@/lib/utils/classnames";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

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
    <Row
      gap={1}
      align="center"
      wrap
      asContainer
      accessibilityRole="tablist"
      className="mb-[4px]"
      {...props}
    >
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
  disabled?: boolean;
  accessibilityLabel?: string;
};

const Tab = ({
  children,
  value,
  leftIcon,
  rightIcon,
  className,
  disabled = false,
  accessibilityLabel,
}: TabProps) => {
  const { activeTab, onTabChange } = useTabsContext();
  const { resolvedTheme } = useTheme();
  const isActive = activeTab === value;
  const isDark = resolvedTheme === "dark";

  return (
    <Pressable
      accessibilityRole="tab"
      accessibilityLabel={accessibilityLabel}
      accessibilityState={{ selected: isActive, disabled }}
      disabled={disabled}
      onPress={() => onTabChange(value)}
      className={cn(
        "min-h-[44px] flex-row items-center justify-center gap-[8px] rounded-full px-[16px] py-[10px]",
        {
          "opacity-40": disabled,
        },
        className,
      )}
      style={({ pressed }) => ({
        opacity: disabled ? 0.4 : pressed ? 0.65 : 1,
        backgroundColor: isActive
          ? themeColors[resolvedTheme].surfaceElevated
          : "transparent",
        boxShadow: isActive
          ? isDark
            ? "0 0 0 1px rgba(255, 255, 255, 0.06), 0 3px 12px rgba(0, 0, 0, 0.18)"
            : "0 3px 16px rgba(0, 0, 0, 0.06)"
          : undefined,
      })}
    >
      {leftIcon}
      <Text
        fontSize="sm"
        fontWeight={isActive ? "medium" : "normal"}
        color={isActive ? undefined : "muted"}
      >
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

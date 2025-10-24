import React from "react";
import { View, ViewProps } from "react-native";
import { cn } from "@/lib/utils";

export interface WrapperProps extends ViewProps {
  children?: React.ReactNode;
}

export const Wrapper = ({ children, className, ...rest }: WrapperProps) => {
  return (
    <View
      className={cn(
        "rounded-2xl border bg-gray-50/10 border-gray-100/70 py-3 px-4 shadow-lg shadow-gray-50 dark:border-dark-200 dark:bg-dark/80 dark:shadow-none",
        className
      )}
      {...rest}
    >
      {children}
    </View>
  );
};

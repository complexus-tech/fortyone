import React from "react";
import { View, ViewProps } from "react-native";
import { cn } from "@/lib/utils/classnames";

export interface WrapperProps extends ViewProps {
  children?: React.ReactNode;
}

export const Wrapper = ({ children, className, ...rest }: WrapperProps) => {
  return (
    <View
      className={cn(
        "rounded-2xl border border-border bg-surface px-4 py-3 dark:border-border-dark dark:bg-surface-dark",
        className,
      )}
      {...rest}
    >
      {children}
    </View>
  );
};

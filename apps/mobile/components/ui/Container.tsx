import { View, ViewProps } from "react-native";
import { cn } from "@/lib/utils";

export interface ContainerProps extends ViewProps {
  children?: React.ReactNode;
}

export const Container = ({
  children,
  className,
  ...props
}: ContainerProps) => {
  return (
    <View className={cn("px-4.5 flex-1", className)} {...props}>
      {children}
    </View>
  );
};

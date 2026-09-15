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
    <View className={cn("flex-1 px-[20px]", className)} {...props}>
      {children}
    </View>
  );
};

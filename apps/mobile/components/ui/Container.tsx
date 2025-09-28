import { View, ViewProps } from "react-native";

export interface ContainerProps extends ViewProps {
  children?: React.ReactNode;
  full?: boolean;
}

export const Container = ({
  children,
  className,
  full,
  ...props
}: ContainerProps) => {
  const paddingClass = full ? "px-0" : "px-6";

  return (
    <View className={`${paddingClass} ${className || ""}`} {...props}>
      {children}
    </View>
  );
};

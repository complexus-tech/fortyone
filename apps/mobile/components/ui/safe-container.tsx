import {
  SafeAreaView,
  SafeAreaViewProps,
} from "react-native-safe-area-context";
import { StyleSheet } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";

export interface ContainerProps extends SafeAreaViewProps {
  children?: React.ReactNode;
  isFull?: boolean;
}

export const SafeContainer = ({
  children,
  isFull,
  style,
  edges = ["top"],
  ...props
}: ContainerProps) => {
  const { resolvedTheme } = useTheme();

  const containerStyle = [
    styles.container,
    isFull && styles.full,
    {
      backgroundColor: themeColors[resolvedTheme].background,
    },
    style,
  ];

  return (
    <SafeAreaView style={containerStyle} edges={edges} {...props}>
      {children}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingHorizontal: 20,
  },
  full: {
    paddingHorizontal: 0,
  },
});

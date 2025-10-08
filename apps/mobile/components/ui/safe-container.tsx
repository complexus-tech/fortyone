import {
  SafeAreaView,
  SafeAreaViewProps,
} from "react-native-safe-area-context";
import { StyleSheet } from "react-native";
import { colors } from "@/constants/colors";
import { useColorScheme } from "nativewind";

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
  const { colorScheme } = useColorScheme();

  const containerStyle = [
    styles.container,
    isFull && styles.full,
    {
      backgroundColor:
        colorScheme === "dark" ? colors.dark.DEFAULT : colors.white,
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
    paddingHorizontal: 18,
  },
  full: {
    paddingHorizontal: 0,
  },
});

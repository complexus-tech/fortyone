import {
  SafeAreaView,
  SafeAreaViewProps,
} from "react-native-safe-area-context";
import { StyleSheet } from "react-native";

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
  return (
    <SafeAreaView
      style={[styles.container, isFull && styles.full, style]}
      edges={edges}
      {...props}
    >
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

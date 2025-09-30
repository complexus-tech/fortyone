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
  ...props
}: ContainerProps) => {
  return (
    <SafeAreaView
      style={[styles.container, isFull && styles.full, style]}
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

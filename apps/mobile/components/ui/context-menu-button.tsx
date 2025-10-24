import React from "react";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
import { Pressable, StyleProp, View, ViewStyle } from "react-native";
import { Ionicons } from "@expo/vector-icons";

type ContextMenuAction = {
  systemImage?: string;
  label: string;
  onPress: () => void;
  color?: string;
};

type ContextMenuButtonProps = {
  actions: ContextMenuAction[];
  children?: React.ReactNode;
  withNoHost?: boolean;
  hostStyle?: StyleProp<ViewStyle>;
};

const Menu = ({ actions, children }: ContextMenuButtonProps) => {
  const { resolvedTheme } = useTheme();

  // For Android, we'll use a simple pressable with basic styling
  // In a real implementation, you might want to use a proper menu library
  return (
    <Pressable
      style={{
        width: 40,
        height: 40,
        borderRadius: 20,
        backgroundColor:
          resolvedTheme === "light"
            ? "rgba(255, 255, 255, 0.8)"
            : "rgba(0, 0, 0, 0.8)",
        justifyContent: "center",
        alignItems: "center",
        shadowColor: "#000",
        shadowOffset: {
          width: 0,
          height: 2,
        },
        shadowOpacity: 0.1,
        shadowRadius: 4,
        elevation: 3,
      }}
      onPress={() => {
        // For Android, we'll trigger the first action for now
        // In a real implementation, you'd show a proper menu
        if (actions.length > 0) {
          actions[0].onPress();
        }
      }}
    >
      {children ? (
        children
      ) : (
        <Ionicons
          name="ellipsis-horizontal"
          size={20}
          color={resolvedTheme === "light" ? colors.dark[50] : colors.gray[200]}
        />
      )}
    </Pressable>
  );
};

export const ContextMenuButton = ({
  actions,
  children,
  withNoHost,
  hostStyle = {
    width: 40,
    height: 40,
  },
}: ContextMenuButtonProps) => {
  if (withNoHost) {
    return (
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    );
  }
  return (
    <View style={hostStyle}>
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    </View>
  );
};

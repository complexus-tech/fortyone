import type { HeaderActionsProps } from "./header-actions.types";
import { StyleSheet, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { ContextMenuButton } from "./context-menu-button";
import { NewStoryIcon } from "@/components/icons/new-story";
import { IconButton } from "./icon-button";

export type { HeaderActionsProps } from "./header-actions.types";

export function HeaderActions({
  onCreate,
  createLabel,
  actions,
  onOptions,
  optionsLabel = "View options",
  menuLabel = "More options",
  menuIcon = "ellipsis-horizontal",
}: HeaderActionsProps) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];

  return (
    <View
      style={[
        styles.capsule,
        {
          backgroundColor: theme.surfaceElevated,
          borderColor: theme.border,
        },
      ]}
    >
      <IconButton label={createLabel} onPress={onCreate} style={styles.action}>
        <NewStoryIcon size={22} color={theme.foreground} />
      </IconButton>
      {onOptions ? (
        <IconButton
          icon="ellipsis-horizontal"
          label={optionsLabel}
          onPress={onOptions}
          style={styles.action}
        />
      ) : null}
      {actions?.length ? (
        <ContextMenuButton actions={actions} menuLabel={menuLabel}>
          <Ionicons
            accessible={false}
            name={menuIcon}
            size={20}
            color={theme.foreground}
          />
        </ContextMenuButton>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  capsule: {
    flexDirection: "row",
    alignItems: "center",
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
    overflow: "hidden",
    boxShadow: "0 2px 8px rgba(0, 0, 0, 0.05)",
  },
  action: {
    borderRadius: 999,
  },
});

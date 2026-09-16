import {
  Alert,
  ActivityIndicator,
  FlatList,
  Modal,
  Pressable,
  StyleSheet,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { SafeContainer, Text } from "@/components/ui";
import { GlassIconButton } from "@/components/ui/glass-icon-button";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";

export function MayaHistory({
  visible,
  onClose,
  sessions,
  loading,
  currentId,
  onSelect,
  onDelete,
}: {
  visible: boolean;
  onClose: () => void;
  sessions: { id: string; title: string }[];
  loading: boolean;
  currentId: string;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
}) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  return (
    <Modal
      visible={visible}
      presentationStyle="pageSheet"
      animationType="slide"
      onRequestClose={onClose}
    >
      <SafeContainer isFull edges={["top", "bottom"]}>
        <View style={styles.header}>
          <Text fontSize="2xl" fontWeight="semibold" style={{ flex: 1 }}>
            Conversations
          </Text>
          <GlassIconButton
            icon="close"
            systemImage="xmark"
            label="Close conversations"
            onPress={onClose}
          />
        </View>
        {loading ? (
          <ActivityIndicator
            color={theme.textMuted}
            style={{ marginTop: 30 }}
          />
        ) : (
          <FlatList
            data={sessions}
            keyExtractor={(item) => item.id}
            contentContainerStyle={{ paddingHorizontal: 16, paddingBottom: 24 }}
            ListEmptyComponent={
              <View style={styles.empty}>
                <Ionicons
                  name="chatbubble-outline"
                  size={28}
                  color={theme.textMuted}
                />
                <Text fontWeight="medium">A fresh start</Text>
                <Text color="muted" align="center">
                  Your conversations with Maya will appear here.
                </Text>
              </View>
            }
            renderItem={({ item }) => (
              <View
                style={[
                  styles.row,
                  {
                    backgroundColor:
                      currentId === item.id
                        ? theme.surfaceMuted
                        : "transparent",
                  },
                ]}
              >
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel={`Open ${item.title || "Untitled conversation"}`}
                  accessibilityState={{ selected: currentId === item.id }}
                  onPress={() => onSelect(item.id)}
                  style={styles.select}
                >
                  <Text numberOfLines={1} style={{ flex: 1 }}>
                    {item.title || "Untitled conversation"}
                  </Text>
                </Pressable>
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel={`Delete ${item.title || "conversation"}`}
                  style={styles.delete}
                  onPress={() =>
                    Alert.alert(
                      "Delete conversation?",
                      "This conversation will be removed from your history.",
                      [
                        { text: "Cancel", style: "cancel" },
                        {
                          text: "Delete",
                          style: "destructive",
                          onPress: () => onDelete(item.id),
                        },
                      ],
                    )
                  }
                >
                  <Ionicons
                    name="trash-outline"
                    size={18}
                    color={theme.textMuted}
                  />
                </Pressable>
              </View>
            )}
          />
        )}
      </SafeContainer>
    </Modal>
  );
}

const styles = StyleSheet.create({
  header: { padding: 20, flexDirection: "row", alignItems: "center", gap: 12 },
  row: {
    flexDirection: "row",
    alignItems: "center",
    borderRadius: 16,
    marginBottom: 4,
  },
  select: {
    flex: 1,
    flexDirection: "row",
    gap: 12,
    alignItems: "center",
    padding: 14,
    minHeight: 58,
  },
  delete: {
    width: 48,
    height: 48,
    alignItems: "center",
    justifyContent: "center",
  },
  empty: { alignItems: "center", padding: 32, paddingTop: 70, gap: 12 },
});

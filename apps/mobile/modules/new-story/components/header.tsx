import {
  ActivityIndicator,
  Pressable,
  View,
  useColorScheme,
} from "react-native";
import { useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import { IconButton, Text } from "@/components/ui";
import { colors, themeColors } from "@/constants/colors";

type HeaderProps = {
  disabled?: boolean;
  loading?: boolean;
  onSubmit: () => void;
  teamName?: string;
  teamColor?: string | null;
  metadataDisabled?: boolean;
  onTeamPress: () => void;
};

export const Header = ({
  disabled,
  loading,
  onSubmit,
  teamName,
  teamColor,
  metadataDisabled,
  onTeamPress,
}: HeaderProps) => {
  const router = useRouter();
  const dark = useColorScheme() === "dark";
  const surface = themeColors[dark ? "dark" : "light"].surfaceMuted;
  const muted = themeColors[dark ? "dark" : "light"].textMuted;
  const submitDisabled = Boolean(disabled || loading);
  return (
    <View
      style={{
        flexDirection: "row",
        alignItems: "center",
        gap: 12,
        paddingHorizontal: 20,
        paddingVertical: 12,
      }}
    >
      <IconButton
        icon="close"
        label="Close task composer"
        disabled={loading}
        onPress={() => {
          if (router.canGoBack()) router.back();
          else router.replace("/");
        }}
        style={{ borderRadius: 22, backgroundColor: surface }}
      />
      <View style={{ flex: 1, minWidth: 0, alignItems: "center" }}>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={`Change team: ${teamName || "Choose team"}`}
          accessibilityState={{ disabled: metadataDisabled }}
          disabled={metadataDisabled}
          onPress={onTeamPress}
          style={({ pressed }) => ({
            maxWidth: "100%",
            minHeight: 44,
            flexDirection: "row",
            alignItems: "center",
            gap: 8,
            paddingHorizontal: 12,
            paddingVertical: 8,
            borderRadius: 22,
            backgroundColor: surface,
            opacity: metadataDisabled ? 0.4 : pressed ? 0.6 : 1,
          })}
        >
          <View
            style={{
              width: 10,
              height: 10,
              borderRadius: 3,
              backgroundColor: teamColor || muted,
            }}
          />
          <Text
            fontSize="sm"
            fontWeight="medium"
            numberOfLines={1}
            style={{ flexShrink: 1 }}
          >
            {teamName || "Choose team"}
          </Text>
          <Ionicons name="chevron-down" size={13} color={muted} />
        </Pressable>
      </View>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={loading ? "Creating task" : "Create task"}
        accessibilityState={{
          disabled: submitDisabled,
          busy: Boolean(loading),
        }}
        disabled={submitDisabled}
        onPress={onSubmit}
        style={({ pressed }) => ({
          width: 44,
          height: 44,
          borderRadius: 22,
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: surface,
          opacity: pressed ? 0.6 : 1,
        })}
      >
        {loading ? (
          <ActivityIndicator color={colors.primary} />
        ) : (
          <Ionicons
            name="arrow-up"
            size={24}
            color={
              submitDisabled
                ? themeColors[dark ? "dark" : "light"].textDisabled
                : themeColors[dark ? "dark" : "light"].foreground
            }
          />
        )}
      </Pressable>
    </View>
  );
};

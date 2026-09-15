import type { TeamSwitcherProps } from "./team-switcher.types";
import { useState } from "react";
import {
  FlatList,
  Modal,
  Pressable,
  StyleSheet,
  TextInput,
  View,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { GlassIconButton } from "./glass-icon-button";
import { ScreenHeader } from "./screen-header";
import { Text } from "./Text";
import { filterSwitcherTeams } from "./team-switcher-utils";
import { useTeamSwitcherNavigation } from "./use-team-switcher-navigation";

export type { TeamSwitcherProps } from "./team-switcher.types";

function TeamSwitcherContent({
  teams,
  currentTeamId,
  onClose,
  onSelect,
}: Pick<TeamSwitcherProps, "teams" | "currentTeamId"> & {
  onClose: () => void;
  onSelect: (id: string) => void;
}) {
  const [query, setQuery] = useState("");
  const { resolvedTheme } = useTheme();
  const dark = resolvedTheme === "dark";
  const foreground = themeColors[dark ? "dark" : "light"].foreground;
  const muted = themeColors[dark ? "dark" : "light"].textMuted;

  return (
    <SafeAreaProvider>
      <SafeAreaView
        style={[
          styles.container,
          { backgroundColor: themeColors[dark ? "dark" : "light"].background },
        ]}
      >
        <ScreenHeader
          title="Your teams"
          compact
          trailing={
            <GlassIconButton
              icon="close"
              systemImage="xmark"
              label="Close team switcher"
              onPress={onClose}
            />
          }
        />
        <View
          style={[
            styles.search,
            {
              backgroundColor:
                themeColors[dark ? "dark" : "light"].surfaceMuted,
            },
          ]}
        >
          <Ionicons name="search" color={muted} size={18} accessible={false} />
          <TextInput
            accessibilityLabel="Search your teams"
            placeholder="Search teams"
            placeholderTextColor={muted}
            value={query}
            onChangeText={setQuery}
            autoCorrect={false}
            autoCapitalize="none"
            returnKeyType="search"
            style={[styles.input, { color: foreground }]}
          />
        </View>
        <FlatList
          data={filterSwitcherTeams(teams, query)}
          keyExtractor={(team) => team.id}
          extraData={currentTeamId}
          keyboardShouldPersistTaps="handled"
          keyboardDismissMode="on-drag"
          contentContainerStyle={{ paddingBottom: 20 }}
          ListEmptyComponent={
            <Text color="muted" style={styles.empty}>
              {teams.length
                ? "No matching teams."
                : "You haven’t joined a team yet."}
            </Text>
          }
          renderItem={({ item: team }) => (
            <Pressable
              accessibilityRole="button"
              accessibilityLabel={`Open ${team.name}, ${team.code}`}
              accessibilityState={{ selected: team.id === currentTeamId }}
              onPress={() => onSelect(team.id)}
              style={({ pressed }) => [
                styles.row,
                { opacity: pressed ? 0.6 : 1 },
              ]}
            >
              <View style={[styles.color, { backgroundColor: team.color }]} />
              <View style={styles.team}>
                <Text numberOfLines={1} ellipsizeMode="tail">
                  {team.name}
                </Text>
                <Text
                  color="muted"
                  fontSize="sm"
                  numberOfLines={1}
                  ellipsizeMode="tail"
                >
                  {team.code}
                </Text>
              </View>
              {team.id === currentTeamId ? (
                <Ionicons
                  name="checkmark"
                  color={foreground}
                  size={18}
                  accessible={false}
                />
              ) : null}
            </Pressable>
          )}
        />
      </SafeAreaView>
    </SafeAreaProvider>
  );
}

export function TeamSwitcher(props: TeamSwitcherProps) {
  const { close, selectTeam } = useTeamSwitcherNavigation(props);
  return (
    <Modal
      visible={props.isOpened}
      presentationStyle="pageSheet"
      animationType="slide"
      onRequestClose={close}
    >
      <TeamSwitcherContent
        {...props}
        onClose={close}
        onSelect={(id) => selectTeam(id, false)}
      />
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, paddingTop: 12 },
  search: {
    marginHorizontal: 20,
    marginBottom: 12,
    paddingHorizontal: 14,
    borderRadius: 24,
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
  input: { flex: 1, minHeight: 48, fontSize: 16, paddingVertical: 12 },
  empty: { padding: 20 },
  row: {
    minHeight: 60,
    paddingHorizontal: 20,
    paddingVertical: 12,
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
  },
  color: { width: 12, height: 12, borderRadius: 3 },
  team: { flex: 1, minWidth: 0, gap: 2 },
});

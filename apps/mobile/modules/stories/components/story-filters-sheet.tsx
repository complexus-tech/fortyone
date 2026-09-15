import type {
  StoryFilterFacet,
  StoryFilterOption,
  StoryFiltersSheetProps,
} from "./story-filters.types";
import { useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Keyboard,
  Modal,
  Pressable,
  StyleSheet,
  TextInput,
  View,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { GlassIconButton } from "@/components/ui/glass-icon-button";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import {
  createTeamStoryFilters,
  getTeamStoryFilterCount,
} from "@/modules/teams/stories/team-story-filters";
import { useStoryFilterSections } from "../hooks/use-story-filter-sections";
import { changeStoryFilter } from "./story-filter-selection";
import { StoryFilterIcon } from "./story-filter-icon";

const FACET_ICONS = {
  sprint: "calendar-outline",
  objective: "locate-outline",
} as const;

export function StoryFiltersSheet(props: StoryFiltersSheetProps) {
  const close = () => {
    Keyboard.dismiss();
    props.onClose();
  };

  return (
    <Modal
      visible={props.isOpen}
      animationType="slide"
      presentationStyle="pageSheet"
      allowSwipeDismissal
      onRequestClose={close}
    >
      {/* Reset navigation/search on every presentation, including external close. */}
      {props.isOpen ? (
        <StoryFiltersContent
          key={`${props.filters.teamId}:${props.initialFacet ?? "root"}`}
          {...props}
          onClose={close}
        />
      ) : null}
    </Modal>
  );
}

function StoryFiltersContent(props: StoryFiltersSheetProps) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const [facet, setFacet] = useState<StoryFilterFacet | null>(
    props.initialFacet ?? null,
  );
  const [query, setQuery] = useState("");
  const sections = useStoryFilterSections(props);
  const section = sections.find(({ id }) => id === facet);
  const search = query.trim().toLowerCase();
  const options = section?.options.filter(
    ({ label, description }) =>
      label.toLowerCase().includes(search) ||
      description?.toLowerCase().includes(search),
  );
  const hasFilters = getTeamStoryFilterCount(props.filters) > 0;
  const openFacet = (next: StoryFilterFacet | null) => {
    Keyboard.dismiss();
    setQuery("");
    setFacet(next);
  };
  const select = (id: string | null) => {
    if (!section) return;
    props.onChange((current) =>
      current.teamId === props.filters.teamId
        ? changeStoryFilter(current, section.id, id)
        : current,
    );
  };

  const renderOption = (option: StoryFilterOption | null) => {
    if (!section) return null;
    const selected = option
      ? section.selectedIds.includes(option.id)
      : section.selectedIds.length === 0;
    const radio =
      section.id === "objective" || !option || option.id === "unassigned";
    const label = option?.label ?? `Any ${section.label.toLowerCase()}`;
    return (
      <Pressable
        accessibilityRole={radio ? "radio" : "checkbox"}
        accessibilityLabel={
          option?.description ? `${label}, ${option.description}` : label
        }
        accessibilityState={{ checked: selected }}
        onPress={() => select(option?.id ?? null)}
        style={({ pressed }) => [
          styles.option,
          { backgroundColor: pressed ? theme.stateHover : undefined },
        ]}
      >
        {option &&
        (section.id === "status" ||
          section.id === "priority" ||
          (section.id === "assignee" && option.id === "unassigned")) ? (
          <StoryFilterIcon facet={section.id} option={option} />
        ) : null}
        <View style={styles.optionText}>
          <Text numberOfLines={1} style={{ color: theme.foreground }}>
            {label}
          </Text>
          {option?.description ? (
            <Text
              numberOfLines={1}
              fontSize="sm"
              style={{ color: theme.textMuted }}
            >
              {option.description}
            </Text>
          ) : null}
        </View>
        <Ionicons
          accessible={false}
          name={
            radio
              ? selected
                ? "radio-button-on"
                : "radio-button-off"
              : selected
                ? "checkbox"
                : "square-outline"
          }
          size={22}
          color={selected ? theme.foreground : theme.icon}
        />
      </Pressable>
    );
  };

  return (
    <SafeAreaProvider>
      <SafeAreaView
        style={[styles.container, { backgroundColor: theme.surface }]}
      >
        <View style={styles.header}>
          {section ? (
            <GlassIconButton
              icon="chevron-back"
              systemImage="chevron.left"
              label="Back to filters"
              onPress={() => openFacet(null)}
            />
          ) : null}
          <Text
            accessibilityRole="header"
            numberOfLines={1}
            fontWeight="semibold"
            style={[styles.title, { color: theme.foreground }]}
          >
            {section?.label ?? "Filters"}
          </Text>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={
              section ? "Done filtering" : "Reset all filters"
            }
            accessibilityState={{ disabled: !section && !hasFilters }}
            disabled={!section && !hasFilters}
            onPress={
              section
                ? props.onClose
                : () => {
                    props.onChange((current) =>
                      current.teamId === props.filters.teamId
                        ? createTeamStoryFilters(current.teamId)
                        : current,
                    );
                  }
            }
            style={({ pressed }) => [
              styles.textButton,
              { opacity: pressed ? 0.6 : 1 },
            ]}
          >
            <Text
              style={{
                color:
                  !section && !hasFilters
                    ? theme.textDisabled
                    : theme.foreground,
              }}
            >
              {section ? "Done" : "Reset all"}
            </Text>
          </Pressable>
          {!section ? (
            <GlassIconButton
              icon="close"
              systemImage="xmark"
              label="Close filters"
              onPress={props.onClose}
            />
          ) : null}
        </View>
        {section ? (
          <>
            <View
              style={[styles.search, { backgroundColor: theme.surfaceMuted }]}
            >
              <Ionicons
                accessible={false}
                name="search"
                size={18}
                color={theme.icon}
              />
              <TextInput
                accessibilityLabel={`Search ${section.label.toLowerCase()}`}
                placeholder={`Search ${section.label.toLowerCase()}`}
                placeholderTextColor={theme.textMuted}
                value={query}
                onChangeText={setQuery}
                autoCorrect={false}
                autoCapitalize="none"
                returnKeyType="search"
                onSubmitEditing={Keyboard.dismiss}
                style={[styles.input, { color: theme.foreground }]}
              />
              {query ? (
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel="Clear search"
                  onPress={() => setQuery("")}
                  style={styles.clear}
                >
                  <Ionicons
                    name="close-circle"
                    size={20}
                    color={theme.icon}
                    accessible={false}
                  />
                </Pressable>
              ) : null}
            </View>
            <FlatList
              key={section.id}
              data={options}
              extraData={section.selectedIds}
              keyExtractor={(option) => option.id}
              renderItem={({ item }) => renderOption(item)}
              keyboardShouldPersistTaps="handled"
              keyboardDismissMode="on-drag"
              contentContainerStyle={styles.listContent}
              ListHeaderComponent={
                <>
                  {renderOption(null)}
                  {section.error ? (
                    <View style={styles.message}>
                      <Text style={{ color: theme.textMuted }}>
                        Couldn’t load all options.
                      </Text>
                      <Pressable
                        accessibilityRole="button"
                        accessibilityLabel={`Retry loading ${section.label.toLowerCase()}`}
                        onPress={section.retry}
                        style={styles.textButton}
                      >
                        <Text style={{ color: theme.foreground }}>
                          Try again
                        </Text>
                      </Pressable>
                    </View>
                  ) : null}
                </>
              }
              ListEmptyComponent={
                section.loading ? (
                  <ActivityIndicator
                    accessibilityLabel="Loading filter options"
                    color={theme.icon}
                    style={styles.message}
                  />
                ) : !section.error ? (
                  <Text style={[styles.message, { color: theme.textMuted }]}>
                    {search ? "No matching options." : "No options available."}
                  </Text>
                ) : null
              }
            />
          </>
        ) : (
          <FlatList
            key="facets"
            data={sections}
            keyExtractor={(item) => item.id}
            contentContainerStyle={styles.listContent}
            renderItem={({ item }) => (
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={`${item.label}, ${item.value}`}
                accessibilityHint="Choose filter values"
                onPress={() => openFacet(item.id)}
                style={({ pressed }) => [
                  styles.option,
                  { backgroundColor: pressed ? theme.stateHover : undefined },
                ]}
              >
                {item.id === "status" ||
                item.id === "priority" ||
                item.id === "assignee" ? (
                  <StoryFilterIcon facet={item.id} />
                ) : (
                  <Ionicons
                    accessible={false}
                    name={FACET_ICONS[item.id]}
                    size={20}
                    color={theme.icon}
                  />
                )}
                <Text
                  numberOfLines={1}
                  style={[styles.facetLabel, { color: theme.foreground }]}
                >
                  {item.label}
                </Text>
                <View style={styles.facetTrailing}>
                  <Text
                    numberOfLines={1}
                    style={[styles.facetValue, { color: theme.textMuted }]}
                  >
                    {item.value}
                  </Text>
                  <Ionicons
                    accessible={false}
                    name="chevron-forward"
                    size={16}
                    color={theme.icon}
                  />
                </View>
              </Pressable>
            )}
          />
        )}
      </SafeAreaView>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, paddingTop: 12 },
  header: {
    minHeight: 56,
    paddingHorizontal: 20,
    marginBottom: 12,
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
  title: { flex: 1, minWidth: 0, fontSize: 22, lineHeight: 28 },
  textButton: {
    minHeight: 44,
    paddingHorizontal: 8,
    justifyContent: "center",
    alignItems: "center",
  },
  search: {
    flexDirection: "row",
    alignItems: "center",
    paddingLeft: 14,
    paddingRight: 4,
    marginHorizontal: 20,
    marginBottom: 12,
    borderRadius: 24,
    gap: 10,
  },
  input: {
    flex: 1,
    minWidth: 0,
    minHeight: 48,
    fontSize: 16,
    paddingVertical: 12,
  },
  clear: {
    width: 44,
    height: 44,
    alignItems: "center",
    justifyContent: "center",
  },
  listContent: { paddingBottom: 24 },
  option: {
    minHeight: 56,
    paddingHorizontal: 20,
    paddingVertical: 14,
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
  },
  optionText: { flex: 1, minWidth: 0, gap: 2 },
  facetLabel: { flex: 1, minWidth: 0 },
  facetTrailing: {
    maxWidth: "45%",
    flexShrink: 1,
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  facetValue: { flexShrink: 1, textAlign: "right" },
  message: {
    paddingHorizontal: 20,
    paddingVertical: 16,
    alignItems: "flex-start",
  },
});

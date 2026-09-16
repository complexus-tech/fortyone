import type { SFSymbol } from "expo-symbols";
import type {
  StoryFilterFacet,
  StoryFiltersSheetProps,
} from "./story-filters.types";
import { useState } from "react";
import {
  BottomSheet,
  Button,
  Group,
  Host,
  HStack,
  Image,
  LazyVStack,
  RNHostView,
  ScrollView,
  Spacer,
  Text,
  TextField,
  VStack,
} from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  autocorrectionDisabled,
  background,
  buttonStyle,
  clipShape,
  contentShape,
  disabled,
  font,
  foregroundStyle,
  frame,
  glassEffect,
  lineLimit,
  padding,
  presentationDetents,
  presentationDragIndicator,
  scrollDismissesKeyboard,
  shapes,
  submitLabel,
  textFieldStyle,
  textInputAutocapitalization,
} from "@expo/ui/swift-ui/modifiers";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import {
  createTeamStoryFilters,
  getTeamStoryFilterCount,
} from "@/modules/teams/stories/team-story-filters";
import { useStoryFilterSections } from "../hooks/use-story-filter-sections";
import { changeStoryFilter } from "./story-filter-selection";
import { StoryFilterIcon } from "./story-filter-icon";
import { WebIcon } from "@/components/icons/web-icon";

function FilterIconButton({
  symbol,
  label,
  onPress,
}: {
  symbol: SFSymbol;
  label: string;
  onPress: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const palette = themeColors[resolvedTheme];
  return (
    <Button
      onPress={onPress}
      modifiers={[buttonStyle("plain"), accessibilityLabel(label)]}
    >
      <HStack
        modifiers={[
          frame({ width: 44, height: 44 }),
          background(palette.stateHover),
          clipShape("circle"),
          contentShape(shapes.circle()),
          glassEffect({
            glass: { variant: "regular", interactive: true },
            shape: "circle",
          }),
        ]}
      >
        <Image systemName={symbol} size={18} color={palette.foreground} />
      </HStack>
    </Button>
  );
}

function FilterContent(props: StoryFiltersSheetProps) {
  const { resolvedTheme } = useTheme();
  const palette = themeColors[resolvedTheme];
  const [facet, setFacet] = useState(props.initialFacet);
  const [query, setQuery] = useState("");
  const sections = useStoryFilterSections(props);
  const section = sections.find(({ id }) => id === facet);
  const options =
    section?.options.filter(({ label, description }) =>
      `${label} ${description ?? ""}`
        .toLowerCase()
        .includes(query.trim().toLowerCase()),
    ) ?? [];
  const select = (id: string | null) => {
    if (section)
      props.onChange((current) =>
        current.teamId === props.filters.teamId
          ? changeStoryFilter(current, section.id, id)
          : current,
      );
  };
  const openSection = (next?: StoryFilterFacet) => {
    setQuery("");
    setFacet(next);
  };
  return (
    <VStack
      alignment="leading"
      spacing={12}
      modifiers={[padding({ horizontal: 20, top: 24, bottom: 8 })]}
    >
      <HStack spacing={12}>
        {section ? (
          <FilterIconButton
            symbol="chevron.left"
            label="All filters"
            onPress={() => openSection()}
          />
        ) : null}
        <Text
          modifiers={[
            font({ textStyle: "title3", weight: "semibold" }),
            foregroundStyle(palette.foreground),
            lineLimit(1),
          ]}
        >
          {section?.label ?? "Filters"}
        </Text>
        <Spacer />
        {!section ? (
          <Button
            onPress={() =>
              props.onChange((current) =>
                current.teamId === props.filters.teamId
                  ? createTeamStoryFilters(current.teamId)
                  : current,
              )
            }
            modifiers={[
              buttonStyle("plain"),
              disabled(getTeamStoryFilterCount(props.filters) === 0),
              accessibilityLabel("Reset all filters"),
            ]}
          >
            <Text
              modifiers={[
                font({ textStyle: "callout" }),
                foregroundStyle(palette.textMuted),
                frame({ minHeight: 44 }),
              ]}
            >
              Reset
            </Text>
          </Button>
        ) : null}
        <FilterIconButton
          symbol="xmark"
          label="Close filters"
          onPress={props.onClose}
        />
      </HStack>
      {section ? (
        <HStack
          key={section.id}
          spacing={10}
          modifiers={[
            padding({ horizontal: 14 }),
            background(palette.stateHover),
            clipShape("capsule"),
            frame({ minHeight: 48 }),
          ]}
        >
          <RNHostView matchContents>
            <WebIcon name="search" size={18} color={palette.icon} />
          </RNHostView>
          <TextField
            placeholder={`Search ${section.label.toLowerCase()}`}
            onTextChange={setQuery}
            modifiers={[
              accessibilityLabel(`Search ${section.label.toLowerCase()}`),
              font({ textStyle: "callout" }),
              foregroundStyle(palette.foreground),
              frame({ minHeight: 48 }),
              textFieldStyle("plain"),
              textInputAutocapitalization("never"),
              autocorrectionDisabled(),
              submitLabel("search"),
            ]}
          />
        </HStack>
      ) : null}
      <ScrollView modifiers={[scrollDismissesKeyboard("interactively")]}>
        <LazyVStack
          alignment="leading"
          spacing={0}
          modifiers={[padding({ bottom: 24 })]}
        >
          {section ? (
            <>
              <Button
                onPress={() => select(null)}
                modifiers={[
                  buttonStyle("plain"),
                  accessibilityLabel(
                    `Any ${section.label.toLowerCase()}${section.selectedIds.length ? "" : ", selected"}`,
                  ),
                ]}
              >
                <HStack
                  modifiers={[
                    frame({ minHeight: 52 }),
                    contentShape(shapes.rectangle()),
                  ]}
                >
                  <Text
                    modifiers={[
                      font({ textStyle: "callout" }),
                      foregroundStyle(palette.textMuted),
                    ]}
                  >
                    {section.id === "assignee"
                      ? "Anyone"
                      : `Any ${section.label.toLowerCase()}`}
                  </Text>
                  <Spacer />
                  {!section.selectedIds.length ? (
                    <RNHostView matchContents>
                      <WebIcon
                        name="check"
                        size={18}
                        color={palette.foreground}
                      />
                    </RNHostView>
                  ) : null}
                </HStack>
              </Button>
              {section.error ? (
                <VStack
                  alignment="leading"
                  spacing={8}
                  modifiers={[padding({ vertical: 12 })]}
                >
                  <Text
                    modifiers={[
                      font({ textStyle: "callout" }),
                      foregroundStyle(palette.textSecondary),
                    ]}
                  >{`Could not load ${section.label.toLowerCase()} options.`}</Text>
                  <Button
                    onPress={section.retry}
                    modifiers={[
                      buttonStyle("plain"),
                      accessibilityLabel(
                        `Retry loading ${section.label.toLowerCase()}`,
                      ),
                    ]}
                  >
                    <Text
                      modifiers={[
                        font({ textStyle: "callout", weight: "medium" }),
                        foregroundStyle(palette.foreground),
                        frame({ minHeight: 44 }),
                      ]}
                    >
                      Try again
                    </Text>
                  </Button>
                </VStack>
              ) : null}
              {options.map((option) => (
                <Button
                  key={option.id}
                  onPress={() => select(option.id)}
                  modifiers={[
                    buttonStyle("plain"),
                    accessibilityLabel(
                      `${option.label}${option.description ? `, ${option.description}` : ""}${section.selectedIds.includes(option.id) ? ", selected" : ""}`,
                    ),
                  ]}
                >
                  <HStack
                    spacing={12}
                    modifiers={[
                      padding({ vertical: 12 }),
                      frame({ minHeight: 52 }),
                      contentShape(shapes.rectangle()),
                    ]}
                  >
                    {section.id === "sprint" ||
                    section.id === "objective" ||
                    section.id === "status" ||
                    section.id === "priority" ||
                    (section.id === "assignee" &&
                      option.id === "unassigned") ? (
                      <RNHostView matchContents>
                        <StoryFilterIcon facet={section.id} option={option} />
                      </RNHostView>
                    ) : null}
                    <VStack alignment="leading" spacing={2}>
                      <Text
                        modifiers={[
                          font({ textStyle: "callout" }),
                          foregroundStyle(palette.foreground),
                          lineLimit(1),
                        ]}
                      >
                        {option.label}
                      </Text>
                      {option.description ? (
                        <Text
                          modifiers={[
                            font({ textStyle: "subheadline" }),
                            foregroundStyle(palette.textMuted),
                            lineLimit(1),
                          ]}
                        >
                          {option.description}
                        </Text>
                      ) : null}
                    </VStack>
                    <Spacer />
                    {section.selectedIds.includes(option.id) ? (
                      <RNHostView matchContents>
                        <WebIcon
                          name="check"
                          size={18}
                          color={palette.foreground}
                        />
                      </RNHostView>
                    ) : null}
                  </HStack>
                </Button>
              ))}
              {!options.length && !section.error ? (
                <Text
                  modifiers={[
                    font({ textStyle: "callout" }),
                    foregroundStyle(palette.textMuted),
                    padding({ vertical: 20 }),
                  ]}
                >
                  {section.loading
                    ? "Loading…"
                    : query.trim()
                      ? "No matching results"
                      : "No options available"}
                </Text>
              ) : null}
            </>
          ) : (
            sections.map((item) => (
              <Button
                key={item.id}
                onPress={() => openSection(item.id)}
                modifiers={[
                  buttonStyle("plain"),
                  accessibilityLabel(`${item.label}, ${item.value}`),
                ]}
              >
                <HStack
                  spacing={12}
                  modifiers={[
                    padding({ vertical: 12 }),
                    frame({ minHeight: 56 }),
                    contentShape(shapes.rectangle()),
                  ]}
                >
                  <RNHostView matchContents>
                    <StoryFilterIcon facet={item.id} />
                  </RNHostView>
                  <Text
                    modifiers={[
                      font({ textStyle: "callout" }),
                      foregroundStyle(palette.foreground),
                    ]}
                  >
                    {item.label}
                  </Text>
                  <Spacer />
                  <HStack spacing={6}>
                    <Text
                      modifiers={[
                        font({ textStyle: "callout" }),
                        foregroundStyle(palette.textMuted),
                        lineLimit(1),
                        frame({ maxWidth: 145, alignment: "trailing" }),
                      ]}
                    >
                      {item.value}
                    </Text>
                    <RNHostView matchContents>
                      <WebIcon
                        name="chevronRight"
                        color={palette.icon}
                        size={12}
                      />
                    </RNHostView>
                  </HStack>
                </HStack>
              </Button>
            ))
          )}
        </LazyVStack>
      </ScrollView>
    </VStack>
  );
}

export function StoryFiltersSheet(props: StoryFiltersSheetProps) {
  const { resolvedTheme } = useTheme();
  const [presentation, setPresentation] = useState(0);
  return (
    <Host
      matchContents
      colorScheme={resolvedTheme}
      style={{ position: "absolute" }}
    >
      <BottomSheet
        isPresented={props.isOpen}
        onIsPresentedChange={(open) => {
          if (!open) props.onClose();
        }}
        onDismiss={() => setPresentation((current) => current + 1)}
      >
        <Group
          modifiers={[
            presentationDetents(["medium", "large"]),
            presentationDragIndicator("visible"),
          ]}
        >
          <FilterContent
            key={`${presentation}:${props.initialFacet ?? "root"}`}
            {...props}
          />
        </Group>
      </BottomSheet>
    </Host>
  );
}

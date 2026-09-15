import type { TeamSwitcherProps } from "./team-switcher.types";
import { useState } from "react";
import {
  BottomSheet,
  Button,
  Group,
  Host,
  HStack,
  Image,
  LazyVStack,
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
  const filteredTeams = filterSwitcherTeams(teams, query);

  return (
    <VStack
      spacing={16}
      alignment="leading"
      modifiers={[padding({ horizontal: 20, top: 24, bottom: 8 })]}
    >
      <HStack>
        <Text
          modifiers={[
            font({ textStyle: "title3", weight: "semibold" }),
            foregroundStyle(foreground),
          ]}
        >
          Your teams
        </Text>
        <Spacer />
        <Button
          onPress={onClose}
          modifiers={[
            buttonStyle("plain"),
            accessibilityLabel("Close team switcher"),
          ]}
        >
          <HStack
            modifiers={[
              frame({ width: 44, height: 44 }),
              background(themeColors[resolvedTheme].stateHover),
              clipShape("circle"),
              contentShape(shapes.circle()),
              glassEffect({
                glass: { variant: "regular", interactive: true },
                shape: "circle",
              }),
            ]}
          >
            <Image systemName="xmark" color={foreground} size={20} />
          </HStack>
        </Button>
      </HStack>
      <HStack
        spacing={10}
        modifiers={[
          padding({ horizontal: 14 }),
          frame({ minHeight: 48 }),
          background(themeColors[resolvedTheme].stateHover),
          clipShape("capsule"),
        ]}
      >
        <Image systemName="magnifyingglass" size={18} color={muted} />
        <TextField
          placeholder="Search teams"
          onTextChange={setQuery}
          modifiers={[
            accessibilityLabel("Search your teams"),
            font({ textStyle: "callout" }),
            foregroundStyle(foreground),
            frame({ minHeight: 48 }),
            textFieldStyle("plain"),
            textInputAutocapitalization("never"),
            autocorrectionDisabled(),
            submitLabel("search"),
          ]}
        />
      </HStack>
      <ScrollView modifiers={[scrollDismissesKeyboard("interactively")]}>
        <LazyVStack
          spacing={0}
          alignment="leading"
          modifiers={[padding({ bottom: 20 })]}
        >
          {filteredTeams.length ? (
            filteredTeams.map((team) => (
              <Button
                key={team.id}
                onPress={() => onSelect(team.id)}
                modifiers={[
                  buttonStyle("plain"),
                  accessibilityLabel(
                    `${team.name}, ${team.code}${team.id === currentTeamId ? ", current team" : ""}`,
                  ),
                ]}
              >
                <HStack
                  spacing={12}
                  modifiers={[
                    padding({ vertical: 12 }),
                    frame({ minHeight: 60 }),
                    contentShape(shapes.rectangle()),
                  ]}
                >
                  <Image
                    systemName="square.fill"
                    color={team.color}
                    size={12}
                  />
                  <VStack spacing={2} alignment="leading">
                    <Text
                      modifiers={[
                        font({ textStyle: "callout" }),
                        foregroundStyle(foreground),
                        lineLimit(1),
                      ]}
                    >
                      {team.name}
                    </Text>
                    <Text
                      modifiers={[
                        font({ textStyle: "subheadline" }),
                        foregroundStyle(muted),
                        lineLimit(1),
                      ]}
                    >
                      {team.code}
                    </Text>
                  </VStack>
                  <Spacer />
                  {team.id === currentTeamId ? (
                    <Image
                      systemName="checkmark"
                      color={foreground}
                      size={18}
                    />
                  ) : null}
                </HStack>
              </Button>
            ))
          ) : (
            <Text
              modifiers={[
                font({ textStyle: "callout" }),
                foregroundStyle(muted),
                padding({ vertical: 20 }),
              ]}
            >
              {teams.length
                ? "No matching teams."
                : "You haven’t joined a team yet."}
            </Text>
          )}
        </LazyVStack>
      </ScrollView>
    </VStack>
  );
}

export function TeamSwitcher(props: TeamSwitcherProps) {
  const [presentationId, setPresentationId] = useState(0);
  const { resolvedTheme } = useTheme();
  const { close, selectTeam, finishSelection } =
    useTeamSwitcherNavigation(props);
  return (
    <Host
      matchContents
      colorScheme={resolvedTheme}
      style={{ position: "absolute" }}
    >
      <BottomSheet
        isPresented={props.isOpened}
        onIsPresentedChange={props.setIsOpened}
        onDismiss={() => {
          // Reset the native TextField and JS filter together, after the
          // keyboard and sheet have fully dismissed.
          setPresentationId((current) => current + 1);
          finishSelection();
        }}
      >
        <Group
          modifiers={[
            presentationDetents(["medium", "large"]),
            presentationDragIndicator("visible"),
          ]}
        >
          <TeamSwitcherContent
            key={presentationId}
            {...props}
            onClose={close}
            onSelect={selectTeam}
          />
        </Group>
      </BottomSheet>
    </Host>
  );
}

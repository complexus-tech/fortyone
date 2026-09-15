import { useRef } from "react";
import { ActivityIndicator, TextInput, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useRouter } from "expo-router";
import { HeaderActions, IconButton, Row, ScreenHeader } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { useTerminology } from "@/hooks/use-terminology";

type HeaderProps = {
  value: string;
  onChangeText: (value: string) => void;
  onSubmit: () => void;
  searching: boolean;
  objectivesEnabled: boolean;
  searchType: "stories" | "objectives";
  setSearchType: (type: "stories" | "objectives") => void;
};

export function SearchHeader({
  value,
  onChangeText,
  onSubmit,
  searching,
  objectivesEnabled,
  searchType,
  setSearchType,
}: HeaderProps) {
  const router = useRouter();
  const { resolvedTheme } = useTheme();
  const { getTermDisplay } = useTerminology();
  const inputRef = useRef<TextInput>(null);
  const theme = themeColors[resolvedTheme];
  const scopes = [
    {
      value: "stories" as const,
      label: getTermDisplay("storyTerm", {
        variant: "plural",
        capitalize: true,
      }),
    },
    {
      value: "objectives" as const,
      label: getTermDisplay("objectiveTerm", {
        variant: "plural",
        capitalize: true,
      }),
    },
  ].filter((scope) => objectivesEnabled || scope.value === "stories");
  const selectedLabel = getTermDisplay(
    searchType === "stories" ? "storyTerm" : "objectiveTerm",
    { variant: "plural" },
  );

  return (
    <View>
      <ScreenHeader
        title="Search"
        trailing={
          <HeaderActions
            createLabel={`Create ${getTermDisplay("storyTerm")}`}
            onCreate={() => router.push("/new")}
            menuLabel={`Search type: ${selectedLabel}`}
            menuSystemImage="line.3.horizontal.decrease"
            menuIcon="filter-outline"
            actions={scopes.map((scope) => ({
              label: scope.label,
              selected: scope.value === searchType,
              onPress: () => setSearchType(scope.value),
            }))}
          />
        }
      />
      <View className="px-[20px] pb-3">
        <Row
          align="center"
          gap={2}
          className="min-h-[48px] rounded-2xl pl-[12px]"
          style={{ backgroundColor: theme.surfaceMuted }}
        >
          <Ionicons name="search-outline" size={20} color={theme.textMuted} />
          <TextInput
            ref={inputRef}
            accessibilityLabel={`Search ${selectedLabel.toLowerCase()}`}
            className="min-h-[48px] flex-1 text-[16px] font-normal"
            style={{ color: theme.foreground }}
            placeholder={`Search ${selectedLabel.toLowerCase()}…`}
            placeholderTextColor={theme.textMuted}
            selectionColor={theme.foreground}
            value={value}
            onChangeText={onChangeText}
            onSubmitEditing={onSubmit}
            maxLength={200}
            returnKeyType="search"
            autoCorrect={false}
            autoCapitalize="none"
            autoFocus
          />
          {searching && (
            <ActivityIndicator
              size="small"
              color={theme.textMuted}
              accessibilityLabel="Searching"
            />
          )}
          {value.length > 0 ? (
            <IconButton
              icon="close-circle"
              label="Clear search"
              onPress={() => {
                onChangeText("");
                inputRef.current?.focus();
              }}
            />
          ) : (
            <View style={{ width: 12 }} />
          )}
        </Row>
      </View>
    </View>
  );
}

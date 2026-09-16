import { useRef } from "react";
import {
  ActivityIndicator,
  Platform,
  StyleSheet,
  TextInput,
  View,
} from "react-native";
import { WebIcon } from "@/components/icons/web-icon";
import { IconButton } from "@/components/ui";
import { GlassInputSurface } from "@/components/ui/glass-input-surface";
import { searchInputStyles } from "@/components/ui/search-input-styles";
import { themeColors } from "@/constants/colors";
import { useTheme, useTerminology } from "@/hooks";

export function SearchInput({
  value,
  onChangeText,
  onSubmit,
  searching,
  searchType,
}: {
  value: string;
  onChangeText: (value: string) => void;
  onSubmit: () => void;
  searching: boolean;
  searchType: "stories" | "objectives";
}) {
  const inputRef = useRef<TextInput>(null);
  const { resolvedTheme } = useTheme();
  const { getTermDisplay } = useTerminology();
  const theme = themeColors[resolvedTheme];
  const label = getTermDisplay(
    searchType === "stories" ? "storyTerm" : "objectiveTerm",
    { variant: "plural" },
  ).toLowerCase();

  return (
    <GlassInputSurface>
      <View style={styles.row}>
        <View pointerEvents="none" style={searchInputStyles.icon}>
          <WebIcon name="search" size={20} color={theme.textMuted} />
        </View>
        <TextInput
          ref={inputRef}
          accessibilityLabel={`Search ${label}`}
          style={[
            searchInputStyles.input,
            styles.input,
            { color: theme.foreground },
          ]}
          placeholder={`Search ${label}…`}
          placeholderTextColor={theme.textMuted}
          selectionColor={theme.foreground}
          keyboardAppearance={resolvedTheme}
          value={value}
          onChangeText={onChangeText}
          onSubmitEditing={onSubmit}
          maxLength={200}
          returnKeyType="search"
          autoCorrect={false}
          autoCapitalize="none"
          autoFocus
        />
        {searching ? (
          <ActivityIndicator
            size="small"
            color={theme.textMuted}
            accessibilityLabel="Searching"
          />
        ) : null}
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
          <View style={{ width: 8 }} />
        )}
      </View>
    </GlassInputSurface>
  );
}

const styles = StyleSheet.create({
  input: {
    // UITextField centers its native font metrics. A custom paragraph line
    // height shifts the glyph baseline down despite textAlignVertical.
    lineHeight: Platform.OS === "ios" ? undefined : 22,
    paddingBottom: Platform.OS === "ios" ? 2 : 0,
  },
  row: {
    minHeight: 48,
    flexDirection: "row",
    alignItems: "center",
    paddingLeft: 16,
    paddingRight: 2,
    gap: 10,
  },
});

import { StyleSheet } from "react-native";

/** Keep single-line search text centered with its sibling icon on both platforms. */
export const searchInputStyles = StyleSheet.create({
  input: {
    flex: 1,
    minWidth: 0,
    minHeight: 48,
    fontSize: 16,
    lineHeight: 22,
    paddingVertical: 0,
    includeFontPadding: false,
    textAlignVertical: "center",
  },
  icon: {
    width: 20,
    height: 22,
    flexShrink: 0,
    alignItems: "center",
    justifyContent: "center",
  },
});

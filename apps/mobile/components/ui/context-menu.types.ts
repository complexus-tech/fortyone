import type { SFSymbol } from "expo-symbols";

export type ContextMenuAction = {
  systemImage?: SFSymbol;
  label: string;
  onPress: () => void;
  color?: string;
  selected?: boolean;
};

import type { ContextMenuAction } from "./context-menu.types";
import type { IconButtonProps } from "./icon-button";
import type { SFSymbol } from "expo-symbols";

export type HeaderActionsProps = {
  onCreate: () => void;
  createLabel: string;
  actions?: ContextMenuAction[];
  onOptions?: () => void;
  optionsLabel?: string;
  menuLabel?: string;
  menuIcon?: IconButtonProps["icon"];
  menuSystemImage?: SFSymbol;
};

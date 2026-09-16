import type { ReactNode } from "react";
import type { ContextMenuAction } from "./context-menu.types";
import type { IconButtonProps } from "./icon-button";
import type { SFSymbol } from "expo-symbols";

export type HeaderActionsProps = {
  onCreate: () => void;
  createLabel: string;
  actions?: ContextMenuAction[];
  onOptions?: () => void;
  optionsLabel?: string;
  optionsIcon?: IconButtonProps["icon"];
  optionsSystemImage?: SFSymbol;
  menuLabel?: string;
  menuContent?: ReactNode;
  menuIcon?: IconButtonProps["icon"];
  menuSystemImage?: SFSymbol;
};

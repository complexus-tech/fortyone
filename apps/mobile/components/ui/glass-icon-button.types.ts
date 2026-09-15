import type { ReactNode } from "react";
import type { SFSymbol } from "expo-symbols";
import type { IconButtonProps } from "./icon-button";

export type GlassIconButtonProps = {
  label: string;
  onPress: () => void;
  disabled?: boolean;
} & (
  | { children: ReactNode; icon?: never; systemImage?: never }
  | {
      children?: never;
      icon: IconButtonProps["icon"];
      systemImage: SFSymbol;
    }
);

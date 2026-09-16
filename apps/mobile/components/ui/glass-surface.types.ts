import type { ViewProps } from "react-native";

export type GlassSurfaceProps = ViewProps & {
  cornerRadius?: number;
  selected?: boolean;
};

import type { VariantProps } from "cva";
import type { textVariants } from "./text-variants";
import { createContext } from "react";

// Composite controls provide defaults without overriding explicit child styles.
export const TextStyleContext = createContext<
  Pick<VariantProps<typeof textVariants>, "color" | "fontWeight"> | undefined
>(undefined);

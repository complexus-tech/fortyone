import type { ReactNode } from "react";
import { Box } from "ui";
import { ProductImage } from "./product";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <Box className="grid h-screen grid-cols-2">
      {children}
      <ProductImage />
    </Box>
  );
}

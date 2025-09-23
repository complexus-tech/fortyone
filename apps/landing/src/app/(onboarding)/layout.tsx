import type { ReactNode } from "react";
import { Box, Flex } from "ui";
import { Blur } from "@/components/ui";
import { ProductImage } from "./product";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <Box className="relative grid h-dvh md:grid-cols-[48%_auto]">
      <Blur className="absolute left-1/2 right-1/2 z-[10] h-[60dvh] w-[80vw] -translate-x-1/2 bg-warning/[0.07] md:hidden" />
      <ProductImage />
      <Flex align="center" className="relative z-[3]" justify="center">
        {children}
      </Flex>
    </Box>
  );
}

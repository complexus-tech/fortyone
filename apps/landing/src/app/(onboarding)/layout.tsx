import type { ReactNode } from "react";
import { Box, Flex } from "ui";
import { Blur } from "@/components/ui";
import { ProductImage } from "./product";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <Box className="relative grid h-dvh md:static md:grid-cols-2">
      <Blur className="absolute left-1/2 right-1/2 z-[10] h-[60dvh] w-[80vw] -translate-x-1/2 bg-warning/[0.07] md:hidden" />
      <Flex
        align="center"
        className="relative z-[3] bg-white dark:bg-black"
        justify="center"
      >
        {children}
      </Flex>
      <ProductImage />
    </Box>
  );
}

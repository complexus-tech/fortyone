import type { ReactNode } from "react";
import { Box, Flex } from "ui";
import { ProductImage } from "@/components/onboarding/product-image";

export const OnboardingLayout = ({ children }: { children: ReactNode }) => {
  return (
    <Box className="relative grid h-dvh md:grid-cols-[48%_auto]">
      <ProductImage />
      <Flex align="center" className="relative z-3" justify="center">
        {children}
      </Flex>
    </Box>
  );
};

import type { ReactNode } from "react";
import { Box } from "ui";
import { ProductImage } from "@/components/onboarding/product-image";

export const OnboardingLayout = ({ children }: { children: ReactNode }) => {
  return (
    <Box className="onboarding-layout relative grid h-dvh md:grid-cols-[48%_auto]">
      <ProductImage />
      <Box className="relative z-3 flex min-h-0 flex-col overflow-y-auto">
        <Box className="my-auto flex w-full justify-center py-8">
          {children}
        </Box>
      </Box>
    </Box>
  );
};

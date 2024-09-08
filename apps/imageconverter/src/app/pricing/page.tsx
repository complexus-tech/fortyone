import { Box } from "ui";
import { Pricing } from "@/components/ui";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Pricing | ImageConveta",
};

export default function Page(): JSX.Element {
  return (
    <Box className="pt-16 md:pt-0">
      <Pricing />
    </Box>
  );
}

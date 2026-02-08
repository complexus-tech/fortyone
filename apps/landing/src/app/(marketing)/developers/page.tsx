import type { Metadata } from "next";
import { Box } from "ui";
import { ComingSoon } from "@/components/ui";

export const metadata: Metadata = {
  title: "Developers | FortyOne",
  description: "Developer resources are coming soon.",
  robots: {
    index: false,
    follow: false,
  },
};

export default function Page() {
  return (
    <Box className="pt-16 md:pt-0">
      <ComingSoon />
    </Box>
  );
}

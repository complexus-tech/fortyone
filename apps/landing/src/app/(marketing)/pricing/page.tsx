import type { Metadata } from "next";
import { Box } from "ui";
import { Pricing } from "@/components/ui";
import { ComparePlans } from "@/components/ui/compare";
import { CallToAction } from "@/components/shared";
import { PricingJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "FortyOne Pricing | Flexible Plans for Teams of All Sizes",
  description:
    "Choose the perfect FortyOne plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
  keywords: [
    "fortyone pricing",
    "project management pricing",
    "OKR software cost",
    "team collaboration plans",
    "project management software pricing",
    "OKR platform cost",
    "enterprise project management",
    "free project management",
  ],
  openGraph: {
    title: "FortyOne Pricing | Flexible Plans for Teams of All Sizes",
    description:
      "Choose the perfect FortyOne plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
  },
  twitter: {
    title: "FortyOne Pricing | Flexible Plans for Teams of All Sizes",
    description:
      "Choose the perfect FortyOne plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
  },
};

export default function Page() {
  return (
    <>
      <PricingJsonLd />
      <Box className="pt-16 md:pt-0">
        <Pricing />
        <ComparePlans />
        <CallToAction />
      </Box>
    </>
  );
}

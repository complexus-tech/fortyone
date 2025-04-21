import type { Metadata } from "next";
import { Box } from "ui";
import { Pricing } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import { ComparePlans } from "@/components/ui/compare";
import { SampleClients } from "@/modules/home";
import { PricingJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Complexus Pricing | Flexible Plans for Teams of All Sizes",
  description:
    "Choose the perfect Complexus plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
  keywords: [
    "complexus pricing",
    "project management pricing",
    "OKR software cost",
    "team collaboration plans",
    "project management software pricing",
    "OKR platform cost",
    "enterprise project management",
    "free project management",
  ],
  openGraph: {
    title: "Complexus Pricing | Flexible Plans for Teams of All Sizes",
    description:
      "Choose the perfect Complexus plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
  },
  twitter: {
    title: "Complexus Pricing | Flexible Plans for Teams of All Sizes",
    description:
      "Choose the perfect Complexus plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
  },
};

export default function Page() {
  return (
    <>
      <PricingJsonLd />
      <Box className="pt-16 md:pt-0">
        <Pricing />
        <SampleClients />
        <ComparePlans />
        <Faqs />
      </Box>
    </>
  );
}

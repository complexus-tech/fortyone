import type { Metadata } from "next";
import { Box } from "ui";
import { Pricing } from "@/components/ui";
import { ComparePlans } from "@/components/ui/compare";
import { CallToAction } from "@/components/shared";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
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
  alternates: {
    canonical: getCanonicalUrl("/pricing"),
  },
  openGraph: {
    title: "FortyOne Pricing | Flexible Plans for Teams of All Sizes",
    description:
      "Choose the perfect FortyOne plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
    url: getCanonicalUrl("/pricing"),
    siteName: "FortyOne",
    type: "website",
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    card: "summary_large_image",
    title: "FortyOne Pricing | Flexible Plans for Teams of All Sizes",
    description:
      "Choose the perfect FortyOne plan for your team. From free starter plans to enterprise solutions, find the right fit for your project management needs.",
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function Page() {
  return (
    <>
      <PricingJsonLd />
      <Box>
        <section className="landing-hero-shell landing-page-frame mt-18 overflow-hidden rounded-[3rem] md:mt-20 md:rounded-[4rem]">
          <Pricing variant="page" />
        </section>
        <ComparePlans />
        <CallToAction />
      </Box>
    </>
  );
}

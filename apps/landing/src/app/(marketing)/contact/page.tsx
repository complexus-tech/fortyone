import type { Metadata } from "next";
import { Hero, Support } from "@/modules/contact";
import { CallToAction } from "@/components/shared";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
import { ContactJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Contact FortyOne | AI Project Management Support",
  description:
    "Contact FortyOne for demos, pricing, implementation support, integrations, or help deciding how AI project management fits your team.",
  keywords: [
    "contact complexus",
    "project management support",
    "OKR support",
    "technical assistance",
    "sales inquiry",
    "customer support",
    "project management demo",
    "complexus demo",
  ],
  alternates: {
    canonical: getCanonicalUrl("/contact"),
  },
  openGraph: {
    title: "Contact FortyOne | AI Project Management Support",
    description:
      "Contact FortyOne for demos, pricing, implementation support, integrations, or help deciding how AI project management fits your team.",
    url: getCanonicalUrl("/contact"),
    siteName: "FortyOne",
    type: "website",
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    card: "summary_large_image",
    title: "Contact FortyOne | AI Project Management Support",
    description:
      "Contact FortyOne for demos, pricing, implementation support, integrations, or help deciding how AI project management fits your team.",
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function Page() {
  return (
    <>
      <ContactJsonLd />
      <Hero />
      <Support />
      <CallToAction />
    </>
  );
}

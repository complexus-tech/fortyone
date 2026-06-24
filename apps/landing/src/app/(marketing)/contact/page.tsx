import type { Metadata } from "next";
import { Hero, Support } from "@/modules/contact";
import { CallToAction } from "@/components/shared";
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
  openGraph: {
    title: "Contact FortyOne | AI Project Management Support",
    description:
      "Contact FortyOne for demos, pricing, implementation support, integrations, or help deciding how AI project management fits your team.",
  },
  twitter: {
    title: "Contact FortyOne | AI Project Management Support",
    description:
      "Contact FortyOne for demos, pricing, implementation support, integrations, or help deciding how AI project management fits your team.",
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

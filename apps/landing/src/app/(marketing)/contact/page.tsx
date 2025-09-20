import type { Metadata } from "next";
import { Hero, Support } from "@/modules/contact";
import { CallToAction } from "@/components/shared";
import { ContactJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Contact FortyOne | Get Support for OKR & Project Management",
  description:
    "Reach out to FortyOne for expert support with your OKR and project management needs. Get in touch with our team for demos, pricing, or technical assistance.",
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
    title: "Contact FortyOne | Get Support for OKR & Project Management",
    description:
      "Reach out to FortyOne for expert support with your OKR and project management needs. Get in touch with our team for demos, pricing, or technical assistance.",
  },
  twitter: {
    title: "Contact FortyOne | Get Support for OKR & Project Management",
    description:
      "Reach out to FortyOne for expert support with your OKR and project management needs. Get in touch with our team for demos, pricing, or technical assistance.",
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

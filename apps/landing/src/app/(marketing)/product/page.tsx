import type { Metadata } from "next";
import { Hero, Features } from "@/modules/product";
import { ProductJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Complexus Features | Complete OKR & Project Management Tools",
  description:
    "Explore Complexus's powerful features: OKR tracking, sprint planning, kanban boards, analytics, and team collaboration tools all in one platform.",
  keywords: [
    "project management features",
    "OKR tracking tools",
    "sprint planning software",
    "team collaboration features",
    "kanban board software",
    "project analytics tools",
    "agile project management",
    "team productivity features",
  ],
  openGraph: {
    title: "Complexus Features | Complete OKR & Project Management Tools",
    description:
      "Explore Complexus's powerful features: OKR tracking, sprint planning, kanban boards, analytics, and team collaboration tools all in one platform.",
  },
  twitter: {
    title: "Complexus Features | Complete OKR & Project Management Tools",
    description:
      "Explore Complexus's powerful features: OKR tracking, sprint planning, kanban boards, analytics, and team collaboration tools all in one platform.",
  },
};

export default function Page() {
  return (
    <>
      <ProductJsonLd />
      <Hero />
      <Features />
    </>
  );
}

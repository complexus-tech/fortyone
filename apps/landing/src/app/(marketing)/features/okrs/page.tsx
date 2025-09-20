import type { Metadata } from "next";
import { CallToAction } from "@/components/shared";
import { Hero, Features } from "@/modules/features/okrs";

export const metadata: Metadata = {
  title: "OKRs | Objectives and Key Results Framework | FortyOne",
  description:
    "Set and achieve strategic goals with FortyOne OKRs. Our powerful framework helps teams track objectives and key results for improved performance and alignment.",
  keywords: [
    "OKRs",
    "objectives and key results",
    "goal tracking",
    "team alignment",
    "performance management",
    "OKR software",
    "strategic planning",
    "goal management",
    "business objectives",
    "key results tracking",
  ],
  openGraph: {
    title: "OKRs | Objectives and Key Results Framework | FortyOne",
    description:
      "Set and achieve strategic goals with FortyOne OKRs. Our powerful framework helps teams track objectives and key results for improved performance and alignment.",
  },
  twitter: {
    title: "OKRs | Objectives and Key Results Framework | FortyOne",
    description:
      "Set and achieve strategic goals with FortyOne OKRs. Our powerful framework helps teams track objectives and key results for improved performance and alignment.",
  },
};

export default function Page() {
  return (
    <>
      <Hero />
      <Features />
      <CallToAction />
    </>
  );
}

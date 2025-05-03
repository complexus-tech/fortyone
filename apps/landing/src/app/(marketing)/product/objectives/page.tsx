import type { Metadata } from "next";
import { CallToAction } from "@/components/shared";
import { Hero, Features } from "@/modules/product/objectives";

export const metadata: Metadata = {
  title: "Objectives | Strategic Goal Setting & Management | Complexus",
  description:
    "Set clear objectives and achieve your strategic goals with Complexus Objectives. Link team efforts to organizational vision and track progress consistently.",
  keywords: [
    "strategic objectives",
    "goal management",
    "objective tracking",
    "team objectives",
    "business goals",
    "company objectives",
    "objective measurement",
    "goal alignment",
    "strategic planning",
    "objective software",
  ],
  openGraph: {
    title: "Objectives | Strategic Goal Setting & Management | Complexus",
    description:
      "Set clear objectives and achieve your strategic goals with Complexus Objectives. Link team efforts to organizational vision and track progress consistently.",
  },
  twitter: {
    title: "Objectives | Strategic Goal Setting & Management | Complexus",
    description:
      "Set clear objectives and achieve your strategic goals with Complexus Objectives. Link team efforts to organizational vision and track progress consistently.",
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

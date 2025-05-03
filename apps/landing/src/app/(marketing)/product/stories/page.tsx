import type { Metadata } from "next";
import { CallToAction } from "@/components/shared";
import { Hero, Features } from "@/modules/product/stories";

export const metadata: Metadata = {
  title: "Stories | Task Management & Project Tracking | Complexus",
  description:
    "Streamline your project management with Complexus Stories. Create, track, and manage tasks efficiently with our intuitive story-based workflow system.",
  keywords: [
    "project management stories",
    "task tracking",
    "user stories",
    "agile stories",
    "project management software",
    "task management",
    "workflow management",
    "project tracking",
  ],
  openGraph: {
    title: "Stories | Task Management & Project Tracking | Complexus",
    description:
      "Streamline your project management with Complexus Stories. Create, track, and manage tasks efficiently with our intuitive story-based workflow system.",
  },
  twitter: {
    title: "Stories | Task Management & Project Tracking | Complexus",
    description:
      "Streamline your project management with Complexus Stories. Create, track, and manage tasks efficiently with our intuitive story-based workflow system.",
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

import type { Metadata } from "next";
import { CallToAction } from "@/components/shared";
import { Hero, Features } from "@/modules/features/sprints";

export const metadata: Metadata = {
  title: "Sprints | Agile Sprint Planning & Management | Complexus",
  description:
    "Plan, execute and track sprints efficiently with Complexus Sprint Management. Streamline your agile development process and boost team productivity.",
  keywords: [
    "sprint planning",
    "agile sprints",
    "scrum sprints",
    "sprint management",
    "sprint tracking",
    "agile development",
    "sprint cycles",
    "development iterations",
    "sprint software",
    "agile team management",
  ],
  openGraph: {
    title: "Sprints | Agile Sprint Planning & Management | Complexus",
    description:
      "Plan, execute and track sprints efficiently with Complexus Sprint Management. Streamline your agile development process and boost team productivity.",
  },
  twitter: {
    title: "Sprints | Agile Sprint Planning & Management | Complexus",
    description:
      "Plan, execute and track sprints efficiently with Complexus Sprint Management. Streamline your agile development process and boost team productivity.",
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

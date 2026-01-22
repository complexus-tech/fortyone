import type { Metadata } from "next";
import { CallToAction } from "@/components/shared";
import { Hero, Features } from "@/modules/features/tasks";

export const metadata: Metadata = {
  title: "Tasks | Task Management & Project Tracking | FortyOne",
  description:
    "Streamline your project management with FortyOne Tasks. Create, track, and manage tasks efficiently with our intuitive task-based workflow system.",
  keywords: [
    "project management tasks",
    "task tracking",
    "team tasks",
    "agile tasks",
    "project management software",
    "task management",
    "workflow management",
    "project tracking",
  ],
  openGraph: {
    title: "Tasks | Task Management & Project Tracking | FortyOne",
    description:
      "Streamline your project management with FortyOne Tasks. Create, track, and manage tasks efficiently with our intuitive task-based workflow system.",
  },
  twitter: {
    title: "Tasks | Task Management & Project Tracking | FortyOne",
    description:
      "Streamline your project management with FortyOne Tasks. Create, track, and manage tasks efficiently with our intuitive task-based workflow system.",
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

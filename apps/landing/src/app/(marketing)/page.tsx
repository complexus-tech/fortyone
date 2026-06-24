import type { Metadata } from "next";
import { CallToAction, JsonLd } from "@/components/shared";
import { Pricing } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  Hero,
  HeroCards,
  SampleClients,
  Integrations,
  CoreValues,
  HowItWorks,
  PlatformWorkflow,
} from "@/modules/home";

export const metadata: Metadata = {
  title: "FortyOne | AI Project Manager for Modern Teams",
  description:
    "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
  keywords: [
    "AI project management",
    "AI project manager",
    "project management platform",
    "team project management",
    "goal tracking software",
    "sprint planning software",
    "task management software",
    "OKR software",
    "team alignment tool",
    "project planning software",
  ],
  openGraph: {
    title: "FortyOne | AI Project Manager for Modern Teams",
    description:
      "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
    url: "/",
    siteName: "FortyOne",
    type: "website",
  },
  twitter: {
    title: "FortyOne | AI Project Manager for Modern Teams",
    description:
      "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
    card: "summary_large_image",
  },
};

export default function Page() {
  return (
    <>
      <JsonLd />
      <Hero />
      <HeroCards />
      <SampleClients />
      <HowItWorks />
      <CoreValues />
      <PlatformWorkflow />
      <Integrations />
      <Pricing className="md:pt-0 md:pb-16" />
      <Faqs />
      <CallToAction />
    </>
  );
}

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
  title: "FortyOne | Customer Feedback and Project Management",
  description:
    "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
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
    title: "FortyOne | Customer Feedback and Project Management",
    description:
      "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
    url: "/",
    siteName: "FortyOne",
    type: "website",
  },
  twitter: {
    title: "FortyOne | Customer Feedback and Project Management",
    description:
      "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
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
      <HeroCards />
      <CoreValues />
      <PlatformWorkflow />
      <Integrations />
      <Pricing className="md:pt-0 md:pb-16" />
      <Faqs />
      <CallToAction />
    </>
  );
}

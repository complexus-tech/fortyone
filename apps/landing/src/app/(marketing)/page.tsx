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
} from "@/modules/home";

export const metadata: Metadata = {
  title: "FortyOne | AI Project Management Platform for Teams",
  description:
    "FortyOne connects goals, sprint plans, tasks, and Maya in one AI project management workspace, so teams can plan faster and spot delivery risks early.",
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
    title: "FortyOne | AI Project Management Platform for Teams",
    description:
      "FortyOne connects goals, sprint plans, tasks, and Maya in one AI project management workspace, so teams can plan faster and spot delivery risks early.",
    url: "/",
    siteName: "FortyOne",
    type: "website",
  },
  twitter: {
    title: "FortyOne | AI Project Management Platform for Teams",
    description:
      "FortyOne connects goals, sprint plans, tasks, and Maya in one AI project management workspace, so teams can plan faster and spot delivery risks early.",
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
      {/* <Features /> */}
      <HowItWorks />
      <CoreValues />
      {/* <Maya /> */}
      <Integrations />
      <Pricing className="md:pt-0 md:pb-16" />
      <Faqs />
      <CallToAction />
    </>
  );
}

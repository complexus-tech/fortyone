import type { Metadata } from "next";
import { CallToAction, JsonLd } from "@/components/shared";
import { Pricing } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  Hero,
  HeroCards,
  SampleClients,
  Integrations,
  Maya,
  CoreValues,
} from "@/modules/home";

export const metadata: Metadata = {
  title: "FortyOne | AI Project Management Platform for Teams",
  description:
    "Keep your team's work, plans, and goals in one place with Maya — your AI project manager. Plan faster, stay aligned, and track progress clearly.",
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
      "Keep your team's work, plans, and goals in one place with Maya — your AI project manager. Plan faster, stay aligned, and track progress clearly.",
    url: "/",
    siteName: "FortyOne",
    type: "website",
  },
  twitter: {
    title: "FortyOne | AI Project Management Platform for Teams",
    description:
      "Keep your team's work, plans, and goals in one place with Maya — your AI project manager. Plan faster, stay aligned, and track progress clearly.",
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
      <CoreValues />
      <Maya />
      <Integrations />
      <Pricing className="md:pt-0 md:pb-16" />
      <Faqs />
      <CallToAction />
    </>
  );
}

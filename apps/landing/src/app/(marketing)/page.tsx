import type { Metadata } from "next";
import { CallToAction, JsonLd } from "@/components/shared";
import { Pricing } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
import {
  CalendarWorkflow,
  DocumentsWorkflow,
  FeedbackWorkflow,
  Hero,
  HowItWorks,
  Integrations,
  MayaWorkflow,
  ProductScreenshot,
  SampleClients,
  StrategyWorkflow,
  Testimonials,
} from "@/modules/home";
import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";

export const metadata: Metadata = {
  title: "FortyOne | Strategy, Feedback, and Project Delivery",
  description:
    "Connect customer feedback, strategic goals, documents, schedules, and daily work in one project plan, with AI support for risk, ownership, and next decisions.",
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
    "customer feedback management",
    "feedback portal",
    "public product roadmap",
    "strategy map software",
    "project documents",
    "work scheduling software",
  ],
  alternates: {
    canonical: getCanonicalUrl("/"),
  },
  openGraph: {
    title: "FortyOne | Strategy, Feedback, and Project Delivery",
    description:
      "Connect customer feedback, strategic goals, documents, schedules, and daily work in one project plan, with AI support for risk, ownership, and next decisions.",
    url: "/",
    siteName: "FortyOne",
    type: "website",
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    title: "FortyOne | Strategy, Feedback, and Project Delivery",
    description:
      "Connect customer feedback, strategic goals, documents, schedules, and daily work in one project plan, with AI support for risk, ownership, and next decisions.",
    card: "summary_large_image",
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function Page() {
  return (
    <>
      <JsonLd />
      <main className="[&_h1]:font-semibold [&_h2]:font-semibold">
        <Hero />
        <ProductScreenshot
          alt="FortyOne My Work board showing tasks grouped by Backlog, To Do, In Progress, and Done"
          cropBrowserOnMobile
          darkImage={myWorkBoardDark}
          lightImage={myWorkBoardLight}
          priority
          url="https://fortyone.app/my-work"
        />
        <SampleClients />
        <HowItWorks />
        <FeedbackWorkflow />
        <Testimonials />
        <StrategyWorkflow />
        <DocumentsWorkflow />
        <CalendarWorkflow />
        <MayaWorkflow />
        <Integrations />
        <Pricing className="md:pt-0 md:pb-16" />
        <Faqs />
        <CallToAction />
      </main>
    </>
  );
}

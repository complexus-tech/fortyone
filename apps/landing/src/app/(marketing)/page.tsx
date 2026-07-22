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
  Hero,
  ProductScreenshot,
  SampleClients,
  Integrations,
  FeedbackWorkflow,
  HowItWorks,
  MayaWorkflow,
  PlatformWorkflow,
  Testimonials,
} from "@/modules/home";
import kanbanImgLight from "../../../public/images/product/kanban-light.webp";
import kanbanImg from "../../../public/images/product/kanban.webp";

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
    "customer feedback management",
    "feedback portal",
    "public product roadmap",
  ],
  alternates: {
    canonical: getCanonicalUrl("/"),
  },
  openGraph: {
    title: "FortyOne | Customer Feedback and Project Management",
    description:
      "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
    url: "/",
    siteName: "FortyOne",
    type: "website",
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    title: "FortyOne | Customer Feedback and Project Management",
    description:
      "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
    card: "summary_large_image",
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function Page() {
  return (
    <>
      <JsonLd />
      <Hero />
      <ProductScreenshot
        alt="FortyOne project board showing planned work, active tasks, and MyAI project guidance"
        darkImage={kanbanImg}
        lightImage={kanbanImgLight}
        priority
        url="complexus.fortyone.app/my-work"
      />
      <SampleClients />
      <HowItWorks />
      <FeedbackWorkflow />
      <Testimonials />
      <MayaWorkflow />
      <PlatformWorkflow />
      <Integrations />
      <Pricing className="md:pt-0 md:pb-16" />
      <Faqs />
      <CallToAction />
    </>
  );
}

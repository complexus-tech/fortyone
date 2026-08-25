import type { Metadata } from "next";
import { cn } from "lib";
import { CallToAction, JsonLd } from "@/components/shared";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
  HOME_METADATA_DESCRIPTION,
  HOME_METADATA_TITLE,
} from "@/lib/seo";
import {
  CustomerStories,
  DecideWhatMattersShowcase,
  FeatureOverview,
  Hero,
  HeroProductScreenshot,
  Integrations,
  ProductWorkflowShowcase,
} from "@/modules/home";
import styles from "./home.module.css";

export const metadata: Metadata = {
  title: HOME_METADATA_TITLE,
  description: HOME_METADATA_DESCRIPTION,
  alternates: {
    canonical: getCanonicalUrl("/"),
  },
  openGraph: {
    title: HOME_METADATA_TITLE,
    description: HOME_METADATA_DESCRIPTION,
    url: "/",
    siteName: "FortyOne",
    type: "website",
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    title: HOME_METADATA_TITLE,
    description: HOME_METADATA_DESCRIPTION,
    card: "summary_large_image",
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function Page() {
  return (
    <>
      <JsonLd />
      <main className="[&_h1]:font-semibold [&_h2]:font-semibold">
        <div
          className={cn(
            "landing-page-frame mt-16 rounded-2xl sm:mt-18 sm:rounded-[3rem] md:mt-20 md:rounded-[4rem]",
            styles.heroShell,
          )}
        >
          <section className="pt-px pb-6 md:pb-10">
            <Hero />
            <HeroProductScreenshot />
          </section>
          <DecideWhatMattersShowcase />
        </div>
        <div className={styles.progressBand}>
          <ProductWorkflowShowcase />
          <CustomerStories />
          <FeatureOverview />
          <Integrations />
        </div>
        <CallToAction className="border-t-0" />
      </main>
    </>
  );
}

import type { Metadata } from "next";
import { cn } from "lib";
import { CallToAction, JsonLd } from "@/components/shared";
import { Faqs } from "@/components/ui/faqs";
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
  Integrations,
  ProductScreenshot,
} from "@/modules/home";
import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";
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
            "landing-page-frame mt-18 rounded-[3rem] md:mt-20 md:rounded-[4rem]",
            styles.heroShell,
          )}
        >
          <section className="pt-px pb-6 md:pb-10">
            <Hero />
            <ProductScreenshot
              alt="FortyOne My Work board showing tasks grouped by Backlog, To Do, In Progress, and Done"
              cropBrowserOnMobile
              darkImage={myWorkBoardDark}
              lightImage={myWorkBoardLight}
              priority
              url="https://fortyone.app/my-work"
            />
          </section>
          <DecideWhatMattersShowcase />
        </div>
        <FeatureOverview />
        <div className={styles.progressBand}>
          <CustomerStories />
          <Integrations />
        </div>
        <Faqs />
        <CallToAction />
      </main>
    </>
  );
}

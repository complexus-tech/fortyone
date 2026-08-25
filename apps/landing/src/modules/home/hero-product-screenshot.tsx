import roadmapTimelineDark from "../../../public/images/product/roadmap-timeline-dark.webp";
import roadmapTimelineLight from "../../../public/images/product/roadmap-timeline-light.webp";
import { ProductScreenshot } from "./product-screenshot";

export const HeroProductScreenshot = () => (
  <section aria-label="FortyOne internal roadmap">
    <ProductScreenshot
      alt="FortyOne internal roadmap showing objectives sequenced across a delivery timeline"
      containerClassName="mt-8 md:mt-10"
      cropBrowserOnMobile
      darkImage={roadmapTimelineDark}
      lightImage={roadmapTimelineLight}
      priority
      reveal={false}
      url="https://fortyone.app/roadmap"
    />
  </section>
);

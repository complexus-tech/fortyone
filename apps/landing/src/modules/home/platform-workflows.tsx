import calendarImageDark from "../../../public/images/product/calendar-week-dark.webp";
import calendarImageLight from "../../../public/images/product/calendar-week-light.webp";
import documentsImageDark from "../../../public/images/product/documents-related-work-dark.webp";
import documentsImageLight from "../../../public/images/product/documents-related-work-light.webp";
import { ProductFeatureSection } from "./product-feature-section";
import { ProductScreenshot } from "./product-screenshot";

export const DocumentsWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Write shared project documents where planning happens, then link each document to the stories and objectives it supports."
      eyebrow="Connected context"
      id="documents"
      title="Keep the brief connected to the work."
    >
      <ProductScreenshot
        alt="FortyOne project brief connected to a delivery story and product objective through Related Work"
        containerClassName="mt-10 md:mt-16"
        darkImage={documentsImageDark}
        lightImage={documentsImageLight}
        url="https://fortyone.app/docs/project-brief"
      />
    </ProductFeatureSection>
  );
};

export const CalendarWorkflow = () => {
  return (
    <ProductFeatureSection
      description="See meetings and scheduled work together, protect focus time, and place assigned work into open windows before the week fills up."
      id="calendar"
      title="Plan work around the time your team actually has."
    >
      <ProductScreenshot
        alt="FortyOne weekly calendar combining team meetings and scheduled project tasks"
        containerClassName="mt-10 md:mt-16"
        darkImage={calendarImageDark}
        lightImage={calendarImageLight}
        url="complexus.fortyone.app/calendar"
      />
    </ProductFeatureSection>
  );
};

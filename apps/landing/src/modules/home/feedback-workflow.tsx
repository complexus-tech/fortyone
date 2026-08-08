import feedbackImageDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackImageLight from "../../../public/images/product/feedback-portal-light.webp";
import { ProductFeatureSection } from "./product-feature-section";
import { ProductScreenshot } from "./product-screenshot";

export const FeedbackWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Give customers one place to submit requests, vote, and follow status. Organise feedback by board, see what rises to the top, and keep roadmap updates beside the original request."
      eyebrow="Customer signal"
      id="feedback"
      title="Collect feedback and show customers what happens next."
    >
      <ProductScreenshot
        alt="Public FortyOne feedback portal showing customer requests, votes, boards, and delivery statuses"
        containerClassName="mt-10 md:mt-16"
        darkImage={feedbackImageDark}
        lightImage={feedbackImageLight}
        url="https://fortyone.app/feedback"
      />
    </ProductFeatureSection>
  );
};

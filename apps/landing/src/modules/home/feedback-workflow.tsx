import { UnderlinedHandwrittenAccent } from "@/components/ui";
import { FeedbackRoadmapSwitcher } from "./feedback-roadmap-switcher";
import { ProductFeatureSection } from "./product-feature-section";

export const FeedbackWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Give customers one place to submit requests, vote, and follow status. Organise feedback by board, see what rises to the top, and keep roadmap updates beside the original request."
      id="feedback"
      title={
        <>
          Collect{" "}
          <UnderlinedHandwrittenAccent tone="danger">
            feedback
          </UnderlinedHandwrittenAccent>{" "}
          and show customers what happens next.
        </>
      }
    >
      <FeedbackRoadmapSwitcher />
    </ProductFeatureSection>
  );
};

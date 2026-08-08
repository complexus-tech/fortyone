import { UnderlinedHandwrittenAccent } from "@/components/ui";
import { ProductFeatureSection } from "./product-feature-section";
import { StrategyRoadmapSwitcher } from "./strategy-roadmap-switcher";

export const StrategyWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Start with the ultimate goal, connect strategic pillars to objectives and key results, then sequence objectives on a roadmap with ownership, health, progress, and dates."
      id="strategy"
      title={
        <>
          Connect{" "}
          <UnderlinedHandwrittenAccent tone="primary">
            strategy
          </UnderlinedHandwrittenAccent>{" "}
          to objectives, roadmaps, and{" "}
          <UnderlinedHandwrittenAccent tone="success">
            daily work
          </UnderlinedHandwrittenAccent>
          .
        </>
      }
    >
      <StrategyRoadmapSwitcher />
    </ProductFeatureSection>
  );
};

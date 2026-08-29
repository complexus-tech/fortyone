import { ProductFeatureSection } from "./product-feature-section";
import { StrategyRoadmapSwitcher } from "./strategy-roadmap-switcher";

export const StrategyWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Start with the ultimate goal, connect strategic pillars to objectives and key results, then sequence objectives on a roadmap with ownership, health, progress, and dates."
      id="strategy"
      title={
        <>
          Connect <span className="text-primary">strategy</span> to objectives,
          roadmaps, and <span className="text-success">daily work</span>.
        </>
      }
    >
      <StrategyRoadmapSwitcher />
    </ProductFeatureSection>
  );
};

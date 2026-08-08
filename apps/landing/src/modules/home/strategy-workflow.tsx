import { ProductFeatureSection } from "./product-feature-section";
import { StrategyRoadmapSwitcher } from "./strategy-roadmap-switcher";

export const StrategyWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Start with the ultimate goal, connect strategic pillars to objectives and key results, then sequence objectives on a roadmap with ownership, health, progress, and dates."
      id="strategy"
      title="Turn purpose into a plan everyone can follow."
    >
      <StrategyRoadmapSwitcher />
    </ProductFeatureSection>
  );
};

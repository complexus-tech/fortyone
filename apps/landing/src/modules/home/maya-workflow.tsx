import mayaImageDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaImageLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import { ProductFeatureSection } from "./product-feature-section";
import { ProductScreenshot } from "./product-screenshot";

export const MayaWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Maya prepares a delivery brief from the work already in FortyOne. Review workload, ownership, and risk, then decide what should change before anything touches the plan."
      id="maya"
      title={
        <>
          Use <span className="text-primary">AI</span> to spot delivery risk
          before work goes off track.
        </>
      }
    >
      <ProductScreenshot
        alt="Maya AI delivery brief summarising team workload, delivery risks, owners, and decisions for the week"
        containerClassName="mt-10 md:mt-16"
        darkImage={mayaImageDark}
        lightImage={mayaImageLight}
        url="https://fortyone.app/maya"
      />
    </ProductFeatureSection>
  );
};

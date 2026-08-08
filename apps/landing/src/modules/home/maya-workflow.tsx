import mayaImageDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaImageLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import { ProductFeatureSection } from "./product-feature-section";
import { ProductScreenshot } from "./product-screenshot";

export const MayaWorkflow = () => {
  return (
    <ProductFeatureSection
      description="Ask Maya to prepare a delivery brief from the work already in FortyOne. Review risk, workload, and ownership, then decide what should change before anything touches the plan."
      id="maya"
      title="Ask Maya where delivery needs attention."
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

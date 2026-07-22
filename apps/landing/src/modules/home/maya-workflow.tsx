import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import mayaImgLight from "../../../public/images/product/maya-task-light.webp";
import mayaImg from "../../../public/images/product/maya-task.webp";
import { ProductScreenshot } from "./product-screenshot";

export const MayaWorkflow = () => {
  return (
    <Box className="py-16 md:py-24" id="maya">
      <Container>
        <Box className="flex flex-col gap-6 md:flex-row md:items-baseline md:justify-between md:gap-16">
          <Box data-landing-reveal>
            <Text as="h2" className="max-w-4xl pb-1 text-4xl md:text-5xl">
              Ask Maya to turn feedback into a task.
            </Text>
          </Box>
          <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
            <Text className="w-full max-w-xl leading-relaxed opacity-70 md:mb-0.5">
              Ask Maya to turn the most requested feedback into a
              ready-to-review task. She reads the request, drafts the scope and
              acceptance criteria, and proposes an owner and estimate. Nothing
              is created until you approve the details.
            </Text>
          </Box>
        </Box>
      </Container>

      <ProductScreenshot
        alt="Maya turning customer feedback into a reviewed task with an owner, estimate, and acceptance criteria"
        containerClassName="mt-10 md:mt-16"
        darkImage={mayaImg}
        lightImage={mayaImgLight}
        priority
        url="complexus.fortyone.app/maya"
      />
    </Box>
  );
};

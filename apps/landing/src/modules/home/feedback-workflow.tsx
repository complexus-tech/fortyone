import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import feedbackImageLight from "../../../public/images/product/feedback-light.webp";
import feedbackImageDark from "../../../public/images/product/feedback.webp";
import { ProductScreenshot } from "./product-screenshot";

export const FeedbackWorkflow = () => {
  return (
    <Box className="py-16 md:py-24" id="feedback">
      <Container>
        <Box className="flex flex-col gap-6 md:flex-row md:items-baseline md:justify-between md:gap-16">
          <Box data-landing-reveal>
            <Text as="h2" className="max-w-4xl pb-1 text-4xl md:text-5xl">
              Collect feedback and show customers what happens next.
            </Text>
          </Box>
          <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
            <Text className="w-full max-w-xl leading-relaxed opacity-70 md:mb-0.5">
              Give customers one place to submit requests and vote on ideas.
              Prioritise what matters, move accepted feedback into the project
              plan, and let customers follow progress on a public roadmap.
            </Text>
          </Box>
        </Box>
      </Container>

      <ProductScreenshot
        alt="FortyOne feedback portal showing customer requests grouped by board and delivery status"
        containerClassName="mt-10 md:mt-16"
        darkImage={feedbackImageDark}
        lightImage={feedbackImageLight}
        url="complexus.fortyone.app/feedback"
      />
    </Box>
  );
};

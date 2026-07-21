import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import kanbanImageLight from "../../../public/images/product/kanban-light.webp";
import kanbanImageDark from "../../../public/images/product/kanban.webp";
import { ProductScreenshot } from "./product-screenshot";

export const MayaWorkflow = () => {
  return (
    <Box className="py-16 md:py-24" id="maya">
      <Container>
        <Box className="flex flex-col gap-6 md:flex-row md:items-baseline md:justify-between md:gap-16">
          <Box data-landing-reveal>
            <Text as="h2" className="max-w-4xl pb-1 text-4xl md:text-5xl">
              Ask Maya to move the work forward.
            </Text>
          </Box>
          <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
            <Text className="w-full max-w-xl leading-relaxed opacity-70 md:mb-0.5">
              Maya reads feedback and project context, spots delivery risks, and
              can create, assign, estimate, schedule, or update work. Review
              important changes before they are applied.
            </Text>
          </Box>
        </Box>
      </Container>

      <ProductScreenshot
        alt="Maya reviewing project work and recommended next actions in FortyOne"
        containerClassName="mt-10 md:mt-16"
        darkImage={kanbanImageDark}
        lightImage={kanbanImageLight}
      />
    </Box>
  );
};

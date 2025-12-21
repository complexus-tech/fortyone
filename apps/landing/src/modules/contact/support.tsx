import { Box, Text, Wrapper } from "ui";
import { Container } from "@/components/ui";

export const Support = () => {
  const features = [
    {
      heading: "Support",
      email: "support@complexus.app",
      description:
        "Get expert assistance with platform setup, technical issues, or general guidance. Our support team is here to ensure your success.",
    },
    {
      email: "sales@complexus.app",
      heading: "Sales",
      description:
        "Discover how complexus can transform your business. Schedule a demo, discuss custom solutions, or learn about enterprise pricing.",
    },
  ];

  return (
    <Container>
      <Box className="relative">
        <Box className="mb-16 grid max-w-4xl grid-cols-1 gap-5 md:mb-10 md:grid-cols-2">
          {features.map(({ heading, description, email }) => (
            <Wrapper className="rounded-3xl py-6 md:py-8" key={heading}>
              <Text align="center" as="h2" fontSize="2xl">
                {heading}
              </Text>
              <Text align="center" className="my-4 opacity-80">
                {description}
              </Text>
              <a
                className="my-1 block text-center text-primary underline underline-offset-4 opacity-90 transition"
                href={`mailto:${email}`}
              >
                {email}
              </a>
            </Wrapper>
          ))}
        </Box>
      </Box>
    </Container>
  );
};

import { Box, Button, Flex, Text } from "ui";
import { EmailIcon, SupportIcon } from "icons";
import { Container } from "@/components/ui";

export const Support = () => {
  const features = [
    {
      heading: "Technical Support",
      email: "support@complexus.app",
      icon: <SupportIcon className="h-16 w-auto" />,
      description:
        "Get expert assistance with platform setup, technical issues, or general guidance. Our support team is here to ensure your success.",
    },
    {
      email: "sales@complexus.app",
      icon: <EmailIcon className="h-16 w-auto" />,
      heading: "Enterprise Sales",
      description:
        "Discover how Complexus can transform your business. Schedule a demo, discuss custom solutions, or learn about enterprise pricing.",
    },
  ];

  return (
    <Container className="max-w-4xl">
      <Box className="relative">
        <Box className="mb-16 grid grid-cols-1 gap-10 md:mb-32 md:grid-cols-2 md:gap-x-20 md:gap-y-16">
          {features.map(({ heading, description, icon, email }, idx) => (
            <Flex align="center" direction="column" key={heading}>
              <Text className="mb-4 text-6xl opacity-30 md:text-8xl">
                {icon}
              </Text>
              <Text
                as="h2"
                fontSize="lg"
                fontWeight="semibold"
                transform="uppercase"
              >
                {heading}
              </Text>
              <a
                className="my-1 inline-block font-semibold opacity-90 transition hover:text-primary"
                href={`mailto:${email}`}
              >
                {email}
              </a>
              <Text align="center" className="my-4 text-lg opacity-80">
                {description}
              </Text>
              <Button href={`mailto:${email}`} rounded="full" size="lg">
                {idx === 0 ? "Contact support" : "Contact sales"}
              </Button>
            </Flex>
          ))}
        </Box>
      </Box>
    </Container>
  );
};

import { Avatar, Box, Button, Flex, Text, Wrapper } from "ui";
import { ArrowRightIcon } from "icons";
import { Container, Blur } from "@/components/ui";

export const Reviews = () => {
  const reviews = [
    {
      id: 1,
      name: "Thomas Davis",
      avatar:
        "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
      content:
        "Caltech employs Complexus to enhance team productivity by 35%, resulting in time savings of 4-6 hours per week per user through its seamless integration with GitHub and simplified agile workflows.",
    },
    {
      id: 2,
      name: "Theresa Webb",
      avatar:
        "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
      content:
        "Caltech employs Complexus to enhance team productivity by 35%, resulting in time savings of 4-6 hours per week per user through its seamless integration with GitHub and simplified agile workflows.",
    },
    {
      id: 3,
      name: "Joseph Mukorivo",
      avatar:
        "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
      content:
        "Caltech employs Complexus to enhance team productivity by 35%, resulting in time savings of 4-6 hours per week per user through its seamless integration with GitHub and simplified agile workflows.",
    },
  ];

  return (
    <Container className="relative py-40">
      <Text
        align="center"
        as="h4"
        className="mb-6 tracking-wider"
        color="primary"
        fontSize="sm"
        fontWeight="medium"
        transform="uppercase"
      >
        Customer stories
      </Text>
      <Text
        align="center"
        as="h3"
        className="mx-auto mb-20 max-w-5xl text-6xl"
        color="gradient"
        fontWeight="medium"
      >
        Unlock the secrets to turbocharge your team&rsquo;s momentum!
      </Text>

      <Box className="grid grid-cols-3 gap-12">
        {reviews.map(({ id, name, avatar, content }) => (
          <Wrapper
            className="animate-gradient h-[38vh] rounded-3xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-[1px] shadow-2xl shadow-secondary/40 dark:border-0"
            key={id}
          >
            <Flex
              className="relative h-full rounded-3xl bg-dark px-8 py-12"
              direction="column"
              justify="between"
            >
              <Box>
                <Flex align="center" className="mb-8" gap={4}>
                  <Avatar name={name} src={avatar} />
                  <Text as="h5" fontSize="lg">
                    {name}
                  </Text>
                </Flex>
                <Text className="mb-5 leading-relaxed">{content}</Text>
              </Box>
              <Button
                className="px-6"
                color="tertiary"
                rightIcon={
                  <ArrowRightIcon className="relative top-[0.5px] h-3.5 w-auto" />
                }
                rounded="full"
              >
                Read more
              </Button>
              <div className="pointer-events-none absolute inset-0 z-10 bg-[url(/noise.png)] opacity-50" />
            </Flex>
          </Wrapper>
        ))}
      </Box>

      <Box className="mx-auto mt-20 w-max">
        <Button
          color="tertiary"
          rightIcon={<ArrowRightIcon className="h-4 w-auto" />}
          rounded="full"
          size="lg"
        >
          See all customer stories
        </Button>
      </Box>

      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[800px] w-[800px] -translate-x-1/2 -translate-y-1/2 bg-primary/40 dark:bg-primary/5" />
    </Container>
  );
};

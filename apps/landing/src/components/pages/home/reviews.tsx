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
        "Complexus is a game-changer! It's intuitive, efficient, and powerful. Say goodbye to wasted time - it's a must-have!",
    },
    {
      id: 2,
      name: "Theresa Webb",
      avatar:
        "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
      content:
        "Complexus simplifies project management like no other system. Highly recommend for enhanced productivity!",
    },
    {
      id: 3,
      name: "Joseph Mukorivo",
      avatar:
        "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
      content:
        "Empowering and effortless - Complexus helps us conquer project complexity. A must-have for achieving exceptional results!",
    },
  ];

  return (
    <Container className="relative pb-16 pt-4 md:pb-32">
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
        className="mx-auto mb-6 max-w-5xl pb-2 text-5xl md:text-7xl"
        color="gradient"
        fontWeight="medium"
      >
        Why they choose complexus.
      </Text>
      <Text
        className="mx-auto mb-10 max-w-2xl text-center text-lg md:text-xl"
        fontWeight="normal"
      >
        Complexus is trusted by some of the world&rsquo;s leading companies.
        Here&rsquo;s what they have to say about their experience.
      </Text>

      <Box className="grid grid-cols-1 gap-6 md:grid-cols-3 md:gap-12">
        {reviews.map(({ id, name, avatar, content }) => (
          <Wrapper
            className="animate-gradient h-[38vh] rounded-2xl shadow-2xl shadow-secondary/20 md:rounded-[1.5rem]"
            key={id}
          >
            <Flex
              className="relative h-full px-4 py-6 md:px-8 md:py-12"
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
            </Flex>
          </Wrapper>
        ))}
      </Box>

      <Box className="mx-auto mt-12 w-max md:mt-20">
        <Button
          color="tertiary"
          rightIcon={<ArrowRightIcon className="h-4 w-auto" />}
          rounded="full"
          size="lg"
        >
          See all customer stories
        </Button>
      </Box>

      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[300px] w-[300px] -translate-x-1/2 -translate-y-1/2 bg-secondary/10 md:h-[900px] md:w-[900px]" />
    </Container>
  );
};

import { Box, BlurImage, Text, Flex, Button } from "ui";
import { Container } from "@/components/ui";

export const Stories = () => {
  const stories = [
    {
      title: "Zambia: Provide meals for children in ECD",
      image: "/images/home/hero.webp",
    },
    {
      title: "Kenya: Take orphaned girls to school",
      image: "/images/home/hero.webp",
    },
    {
      title:
        "Zimbabwe: Support skills training for people living with disabilities ",
      image: "/images/home/hero.webp",
    },
  ];

  return (
    <Box>
      <Box className="grid grid-cols-3">
        {stories.map((s, idx) => (
          <Box className="relative" key={idx}>
            <BlurImage className="aspect-square" quality={100} src={s.image} />
            <Box className="absolute inset-0 flex flex-col justify-end bg-black/20 px-6 py-8">
              <Text
                as="h2"
                className="w-full text-[1.7rem] font-semibold leading-snug text-white"
              >
                {s.title}
              </Text>
            </Box>
          </Box>
        ))}
      </Box>
      <Container className="mt-6">
        <Flex justify="end">
          <Button color="secondary" size="lg">
            Read more of our stories
          </Button>
        </Flex>
      </Container>
    </Box>
  );
};

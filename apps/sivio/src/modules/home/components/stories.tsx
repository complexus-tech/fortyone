import { Box, BlurImage, Text } from "ui";

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
    <Box className="grid grid-cols-3">
      <Box className="relative">
        <BlurImage
          className="aspect-square"
          quality={100}
          src="/images/home/hero.webp"
        />
        <Box className="absolute inset-0 flex flex-col justify-end bg-black/20 p-10">
          <Text as="h2" className="w-full text-2xl font-semibold text-white">
            Zambia: Provide meals for children in ECD
          </Text>
        </Box>
      </Box>
    </Box>
  );
};

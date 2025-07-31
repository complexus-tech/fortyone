import { Box, Text, BlurImage } from "ui";
import Link from "next/link";

type Story = {
  id: number;
  title: string;
  description: string;
  image: string;
  link: string;
};

type StoryCardProps = {
  story: Story;
};

export const StoryCard = ({ story }: StoryCardProps) => {
  return (
    <Box className="border border-dark">
      <BlurImage
        className="aspect-[5/3]"
        quality={100}
        src="/images/about/3.jpg"
      />
      <Box className="p-4">
        <Text as="h3" className="mt-4 text-xl font-bold text-black">
          {story.title}
        </Text>
        <Text className="my-3 text-lg">{story.description}</Text>
        <Link
          className="mt-auto text-sm font-semibold uppercase text-black hover:text-primary"
          href={story.link}
        >
          READ MORE
        </Link>
      </Box>
    </Box>
  );
};

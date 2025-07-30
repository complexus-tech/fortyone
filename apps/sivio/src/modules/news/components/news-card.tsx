import { Box, Text, BlurImage } from "ui";
import Link from "next/link";

type News = {
  id: number;
  title: string;
  description: string;
  image: string;
  link: string;
};

type NewsCardProps = {
  news: News;
};

export const NewsCard = ({ news }: NewsCardProps) => {
  return (
    <Box>
      <BlurImage
        className="aspect-[5/3] rounded-2xl"
        quality={100}
        src="/images/about/3.jpg"
      />
      <Text as="h3" className="mt-4 text-xl font-bold text-black">
        {news.title}
      </Text>
      <Text className="my-3 text-lg">{news.description}</Text>
      <Link
        className="mt-auto text-sm font-semibold uppercase text-black hover:text-primary"
        href={news.link}
      >
        READ MORE
      </Link>
    </Box>
  );
};

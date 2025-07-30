import { Box, Text, BlurImage } from "ui";

type ResourceCardProps = {
  title: string;
  subtitle?: string;
  smallText?: string;
  image: string;
};

export const ResourceCard = ({
  title,
  subtitle,
  smallText,
  image,
}: ResourceCardProps) => {
  return (
    <Box className="group relative h-64 overflow-hidden border border-black">
      <BlurImage alt={title} className="h-full w-full" src={image} />
      <Box className="absolute inset-0 flex flex-col justify-end bg-black/50 p-6">
        <Text className="mb-2 text-3xl font-bold text-white">{title}</Text>
        {subtitle ? (
          <Text className="mb-2 text-lg text-white">{subtitle}</Text>
        ) : null}
        {smallText ? <Text className="text-white">{smallText}</Text> : null}
      </Box>
    </Box>
  );
};

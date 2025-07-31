import { Box, Text, BlurImage } from "ui";
import Link from "next/link";

type Report = {
  id: number;
  title: string;
  description: string;
  image: string;
  link: string;
};

type ReportCardProps = {
  report: Report;
};

export const ReportCard = ({ report }: ReportCardProps) => {
  return (
    <Box>
      <BlurImage
        className="aspect-[5/3] rounded-2xl"
        quality={100}
        src="/images/about/3.jpg"
      />
      <Text as="h3" className="mt-4 text-xl font-bold text-black">
        {report.title}
      </Text>
      <Text className="my-3 text-lg">{report.description}</Text>
      <Link
        className="mt-auto text-sm font-semibold uppercase text-black hover:text-primary"
        href={report.link}
      >
        READ MORE
      </Link>
    </Box>
  );
};

import ky from "ky";

export type LinkMetadata = {
  title?: string;
  description?: string;
  image?: string;
};

export const getLinkMetadata = async (url: string): Promise<LinkMetadata> => {
  try {
    // Validate URL
    const validUrl = new URL(url);
    const response = await ky
      .get(`/api/metadata?url=${encodeURIComponent(validUrl.href)}`, {
        next: {
          revalidate: 60 * 60 * 24, // 24 hours
        },
      })
      .json<LinkMetadata>();

    return response;
  } catch (error) {
    // Return empty metadata on error
    return {
      title: undefined,
      description: undefined,
      image: undefined,
    };
  }
};

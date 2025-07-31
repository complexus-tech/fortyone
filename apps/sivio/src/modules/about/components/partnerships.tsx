import { Box, Text } from "ui";
import Image from "next/image";
import { Container } from "@/components/ui";
import educationSvg from "../../../../public/images/home/education.svg";

export const Partnerships = () => {
  return (
    <Container className="py-28">
      <Box className="grid grid-cols-1 gap-20 md:grid-cols-2">
        <Box>
          <Text
            as="h2"
            className="text-6xl font-black leading-tight text-black"
          >
            Formalizing
            <br />
            Partnerships
            <br />
            Through
            <br />
            Transparency
          </Text>
          <Text as="h3" className="mt-6 max-w-lg text-3xl font-bold text-black">
            Compliance Score & Assessment Criteria
          </Text>
          <Box className="mt-8 flex max-w-2xl flex-col gap-4">
            <Text className="text-lg">
              Before being listed on AfricaGiving, organisations must review and
              digitally sign a Memorandum of Understanding (MOU) with SIVIO
              Institute, the parent organisation overseeing the platform. This
              MOU serves as a formal agreement outlining the terms and
              conditions of use, ensuring clarity on roles, responsibilities,
              and expectations for all listed entities.
            </Text>
            <Text className="text-lg">
              To streamline the onboarding process, organisations must sign the
              MOU digitally before their profiles become active. This ensures
              efficiency, security, and ease of access while reinforcing a
              commitment to AfricaGiving&apos;s principles.
            </Text>
            <Text className="text-lg">
              Organisations undergo a structured evaluation based on five key
              areas. Each criterion contributes to an overall compliance score,
              which reflects the organisation&apos;s level of transparency and
              reliability.
            </Text>
          </Box>
        </Box>

        <Box className="flex flex-col gap-10">
          <Box className="flex gap-4">
            <Image
              alt="Reporting Practices"
              className="h-16 w-auto flex-shrink-0"
              src={educationSvg}
            />
            <Box className="flex flex-col gap-4">
              <Text as="h3" className="text-3xl font-bold text-black">
                Reporting Practices
              </Text>
              <Text className="text-lg">
                Organisations must consistently communicate their activities,
                impact, and use of funds. This includes project reports, impact
                statements, and regular updates shared with donors and
                stakeholders.
              </Text>
            </Box>
          </Box>

          <Box className="flex gap-4">
            <Image
              alt="Accessibility"
              className="h-16 w-auto flex-shrink-0"
              src={educationSvg}
            />
            <Box className="flex flex-col gap-4">
              <Text as="h3" className="text-3xl font-bold text-black">
                Accessibility
              </Text>
              <Text className="text-lg">
                To be visible and accessible to donors, organisations must have
                a functional website, email, or social media presence to engage
                with supporters, and clear channels for communication and
                transparency.
              </Text>
            </Box>
          </Box>

          <Box className="flex gap-4">
            <Image
              alt="Governance Structures"
              className="h-16 w-auto flex-shrink-0"
              src={educationSvg}
            />
            <Box className="flex flex-col gap-4">
              <Text as="h3" className="text-3xl font-bold text-black">
                Governance Structures
              </Text>
              <Text className="text-lg">
                Strong governance ensures responsible oversight and ethical
                decision-making. Organisations must have:
                <ul className="mt-2 list-disc pl-8">
                  <li>
                    A functioning board of directors or advisory group
                    overseeing operations
                  </li>
                  <li>
                    Established leadership structures that support
                    accountability and strategic direction
                  </li>
                </ul>
              </Text>
            </Box>
          </Box>

          <Box className="flex gap-4">
            <Image
              alt="Governance Structures"
              className="h-16 w-auto flex-shrink-0"
              src={educationSvg}
            />
            <Box className="flex flex-col gap-4">
              <Text as="h3" className="text-3xl font-bold text-black">
                Legal Registration
              </Text>
              <Text className="text-lg">
                Organisations must provide proof of official registration with
                the relevant authorities in their country.
              </Text>
            </Box>
          </Box>

          <Box className="flex gap-4">
            <Image
              alt="Governance Structures"
              className="h-16 w-auto flex-shrink-0"
              src={educationSvg}
            />
            <Box className="flex flex-col gap-4">
              <Text as="h3" className="text-3xl font-bold text-black">
                Financial Transparency
              </Text>
              <Text className="text-lg">
                Organisations are assessed on their ability to share: Audited
                financial statements, Annual reports, Basic financial
                disclosures detailing income sources and expenditures.
              </Text>
            </Box>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};

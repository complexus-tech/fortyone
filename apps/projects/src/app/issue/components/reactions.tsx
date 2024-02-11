import {Badge, Flex} from 'ui';
import {SmilePlus} from 'lucide-react';

export const Reactions = () => {
  return <Flex className="mb-4 mt-6" gap={1}>
  <Badge
    color="tertiary"
    rounded="lg"
    size="lg"
    variant="outline"
  >
    👍🏼 1
  </Badge>
  <Badge
    color="tertiary"
    rounded="lg"
    size="lg"
    variant="outline"
  >
    🇿🇼 3
  </Badge>
  <Badge
    color="tertiary"
    rounded="lg"
    size="lg"
    variant="outline"
  >
    👌 2
  </Badge>
  <Badge
    className="px-2"
    color="tertiary"
    rounded="lg"
    size="lg"
    variant="outline"
  >
    <SmilePlus className="h-5 w-auto opacity-80" />
  </Badge>
</Flex>
};

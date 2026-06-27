import { Badge } from "@/components/ui/badge";
import { IconProxy } from "@/components/ui/icon-proxy";

export function ProjectBadge({ projectKey }: { projectKey: string }) {
  return (
    <Badge
      variant="project"
      icon={<IconProxy name="folder" size={15} strokeWidth={2.1} />}
    >
      {projectKey}
    </Badge>
  );
}

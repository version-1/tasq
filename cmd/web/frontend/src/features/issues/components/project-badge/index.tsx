import { Badge } from "@/components/ui/badge";
import { IconProxy } from "@/components/ui/icon-proxy";
import styles from "./index.module.css";

export function ProjectBadge({
  projectKey,
  size = "default",
}: {
  projectKey: string;
  size?: "default" | "small";
}) {
  return (
    <Badge
      className={size === "small" ? styles.small : undefined}
      variant="project"
      icon={<IconProxy name="folder" size={size === "small" ? 12 : 15} strokeWidth={2.1} />}
    >
      {projectKey}
    </Badge>
  );
}

"use client";

import type { ProjectWorkflow } from "@/lib/types";
import { FrontmatterSection } from "./frontmatter-section";
import { WorkflowBodySection } from "./workflow-body-section";
import { WorkflowMetaPanel } from "./workflow-meta-panel";
import styles from "./index.module.css";

export function WorkflowSettingsView({ workflow }: { workflow: ProjectWorkflow }) {
  return (
    <section className={styles.layout}>
      <WorkflowMetaPanel updatedAt={workflow.updatedAt} />
      <FrontmatterSection frontmatter={workflow.frontmatter} />
      <WorkflowBodySection body={workflow.body} />
    </section>
  );
}

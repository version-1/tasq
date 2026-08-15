import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import type { Artifact } from "@/lib/types";
import { pullRequestArtifact } from "@/features/issues/artifacts";
import styles from "./index.module.css";

export function ArtifactsSection({ artifacts }: { artifacts: readonly Artifact[] }) {
  const { t } = useTranslation();
  const pullRequest = pullRequestArtifact(artifacts);

  if (!pullRequest) {
    return null;
  }

  return (
    <section className={styles.section} aria-labelledby="artifacts-heading">
      <h2 id="artifacts-heading" className={styles.heading}>{t("issues.detailPage.artifacts")}</h2>
      <a
        aria-label={t("issues.detailPage.pullRequest")}
        className={styles.link}
        href={pullRequest.data_value}
        rel="noopener noreferrer"
        target="_blank"
      >
        <span className={styles.iconFrame}>
          <IconProxy name="git-pull-request" size={18} strokeWidth={1.8} />
        </span>
        <span className={styles.linkCopy}>
          <span className={styles.linkLabel}>{t("issues.detailPage.pullRequest")}</span>
          <span className={styles.reference}>{pullRequestReference(pullRequest.data_value)}</span>
        </span>
      </a>
    </section>
  );
}

function pullRequestReference(value: string): string {
  try {
    const url = new URL(value);
    const match = url.pathname.match(/^\/([^/]+)\/([^/]+)\/pull\/(\d+)\/?$/);
    return match ? `${match[1]}/${match[2]} #${match[3]}` : url.hostname;
  } catch {
    return value;
  }
}

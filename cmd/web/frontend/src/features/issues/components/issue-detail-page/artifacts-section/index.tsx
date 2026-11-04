import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import { Markdown } from "@/components/ui/markdown";
import { fetchDocumentArtifact, fetchDocumentArtifacts } from "@/lib/api";
import type { Artifact, DocumentArtifact } from "@/lib/types";
import { pullRequestArtifact } from "@/features/issues/artifacts";
import styles from "./index.module.css";

export function ArtifactsSection({ artifacts, issueID }: { artifacts: readonly Artifact[]; issueID: number }) {
  const { t } = useTranslation();
  const pullRequest = pullRequestArtifact(artifacts);
  const [documents, setDocuments] = useState<DocumentArtifact[]>([]);
  const [selectedDocument, setSelectedDocument] = useState<DocumentArtifact | null>(null);
  const [error, setError] = useState("");
  useEffect(() => { let active = true; fetchDocumentArtifacts(issueID, { silent: true }).then((items) => { if (active) setDocuments(items); }).catch((reason) => { if (active) setError(reason instanceof Error ? reason.message : t("issues.detailPage.failedToLoadDocuments")); }); return () => { active = false; }; }, [issueID, t]);

  if (!pullRequest && documents.length === 0 && !error) return null;

  return (
    <section className={styles.section} aria-labelledby="artifacts-heading">
      <h2 id="artifacts-heading" className={styles.heading}>{t("issues.detailPage.artifacts")}</h2>
      {pullRequest && <a
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
      </a>}
      {documents.length > 0 && <div className={styles.documents}><h3 className={styles.documentsHeading}>{t("issues.detailPage.documents")}</h3>{documents.map((document) => <button className={styles.documentButton} key={document.id} onClick={() => fetchDocumentArtifact(issueID, document.id, { silent: true }).then(setSelectedDocument).catch((reason) => setError(reason instanceof Error ? reason.message : t("issues.detailPage.failedToLoadDocuments")))} type="button"><span>{document.title}</span><span>{document.type === "work_report" ? t("issues.detailPage.workReport") : t("issues.detailPage.referenceDocument")}</span></button>)}</div>}
      {selectedDocument && <div className={styles.documentBody}><h3>{selectedDocument.title}</h3><Markdown content={selectedDocument.body ?? ""} emptyText="" /></div>}
      {error && <p className={styles.error} role="status">{error}</p>}
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

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
  const [isLoadingDocuments, setIsLoadingDocuments] = useState(true);
  const [loadingDocumentID, setLoadingDocumentID] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setIsLoadingDocuments(true);
    setError("");
    setSelectedDocument(null);
    fetchDocumentArtifacts(issueID, { silent: true })
      .then((items) => {
        if (active) setDocuments(items);
      })
      .catch((reason) => {
        if (active) setError(reason instanceof Error ? reason.message : t("issues.detailPage.failedToLoadDocuments"));
      })
      .finally(() => {
        if (active) setIsLoadingDocuments(false);
      });
    return () => { active = false; };
  }, [issueID, t]);

  if (!pullRequest && !isLoadingDocuments && documents.length === 0 && !error) return null;

  const selectDocument = (document: DocumentArtifact) => {
    setLoadingDocumentID(document.id);
    setError("");
    fetchDocumentArtifact(issueID, document.id, { silent: true })
      .then(setSelectedDocument)
      .catch((reason) => setError(reason instanceof Error ? reason.message : t("issues.detailPage.failedToLoadDocuments")))
      .finally(() => setLoadingDocumentID(null));
  };

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
      {isLoadingDocuments ? <p className={styles.muted} role="status">{t("issues.detailPage.loadingDocuments")}</p> : null}
      {!isLoadingDocuments && documents.length > 0 ? <div className={styles.documents}><h3 className={styles.documentsHeading}>{t("issues.detailPage.documents")}</h3>{documents.map((document) => <button aria-busy={loadingDocumentID === document.id} className={styles.documentButton} disabled={loadingDocumentID !== null} key={document.id} onClick={() => selectDocument(document)} type="button"><span>{document.title}</span><span>{document.type === "work_report" ? t("issues.detailPage.workReport") : t("issues.detailPage.referenceDocument")}</span></button>)}</div> : null}
      {pullRequest && !isLoadingDocuments && documents.length === 0 && !error ? <p className={styles.muted}>{t("issues.detailPage.noDocuments")}</p> : null}
      {loadingDocumentID ? <p className={styles.muted} role="status">{t("issues.detailPage.loadingDocument")}</p> : null}
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

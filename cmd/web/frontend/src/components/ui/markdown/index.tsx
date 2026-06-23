"use client";

import ReactMarkdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import styles from "./index.module.css";

type MarkdownProps = {
  content: string;
  emptyText: string;
  className?: string;
};

export function Markdown({ content, emptyText, className }: MarkdownProps) {
  const markdown = rewriteAttachmentURLs(content.trim());
  if (markdown === "") {
    return <p className={className}>{emptyText}</p>;
  }
  return (
    <div className={[styles.markdown, className].filter(Boolean).join(" ")}>
      <ReactMarkdown
        components={{
          table: ({ children }) => <table className={styles.table}>{children}</table>,
          td: ({ children }) => <td className={styles.tableCell}>{children}</td>,
          th: ({ children }) => <th className={styles.tableHeader}>{children}</th>,
        }}
        rehypePlugins={[rehypeSanitize]}
        remarkPlugins={[remarkGfm]}
      >
        {markdown}
      </ReactMarkdown>
    </div>
  );
}

function rewriteAttachmentURLs(content: string): string {
  return content.replace(/attachment:\/\/([A-Za-z0-9_-]+)/g, (_, id: string) => {
    return `${issueTrackerURL()}/api/v1/attachments/${id}/content`;
  });
}

function issueTrackerURL(): string {
  return ("/tracker").replace(/\/$/, "");
}

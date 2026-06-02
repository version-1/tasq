"use client";

import ReactMarkdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";

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
    <div className={className}>
      <ReactMarkdown rehypePlugins={[rehypeSanitize]}>{markdown}</ReactMarkdown>
    </div>
  );
}

function rewriteAttachmentURLs(content: string): string {
  return content.replace(/attachment:\/\/([A-Za-z0-9_-]+)/g, (_, id: string) => {
    return `${issueTrackerURL()}/api/v1/attachments/${id}/content`;
  });
}

function issueTrackerURL(): string {
  return (process.env.NEXT_PUBLIC_ISSUE_TRACKER_URL ?? "").replace(/\/$/, "");
}

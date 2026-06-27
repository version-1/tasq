"use client";

import { useEffect, useState, type CSSProperties, type ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";
import { createBundledHighlighter, createSingletonShorthands } from "shiki/core";
import styles from "./index.module.css";

type MarkdownProps = {
  content: string;
  emptyText: string;
  className?: string;
};

type ShikiToken = {
  content: string;
  color?: string;
  fontStyle?: number;
};

type SupportedShikiLanguage = "bash" | "diff" | "json" | "jsonc" | "shell" | "ts" | "tsx" | "typescript";
type SupportedShikiTheme = "github-light";

const createMarkdownHighlighter = createBundledHighlighter<SupportedShikiLanguage, SupportedShikiTheme>({
  engine: () => createJavaScriptRegexEngine(),
  langs: {
    bash: () => import("shiki/dist/langs/bash.mjs"),
    diff: () => import("shiki/dist/langs/diff.mjs"),
    json: () => import("shiki/dist/langs/json.mjs"),
    jsonc: () => import("shiki/dist/langs/jsonc.mjs"),
    shell: () => import("shiki/dist/langs/shell.mjs"),
    ts: () => import("shiki/dist/langs/ts.mjs"),
    tsx: () => import("shiki/dist/langs/tsx.mjs"),
    typescript: () => import("shiki/dist/langs/typescript.mjs"),
  },
  themes: {
    "github-light": () => import("shiki/dist/themes/github-light.mjs"),
  },
});
const { codeToTokens } = createSingletonShorthands(createMarkdownHighlighter);

export function Markdown({ content, emptyText, className }: MarkdownProps) {
  const markdown = rewriteAttachmentURLs(content.trim());
  if (markdown === "") {
    return <p className={className}>{emptyText}</p>;
  }
  return (
    <div className={[styles.markdown, className].filter(Boolean).join(" ")}>
      <ReactMarkdown
        components={{
          a: ({ children, href }) => (
            <a className={styles.link} href={href}>
              {children}
            </a>
          ),
          blockquote: ({ children }) => <blockquote className={styles.blockquote}>{children}</blockquote>,
          code: ({ children, className }) => <MarkdownCode className={className}>{children}</MarkdownCode>,
          h1: ({ children }) => <h1 className={styles.headingPrimary}>{children}</h1>,
          h2: ({ children }) => <h2 className={styles.headingSecondary}>{children}</h2>,
          h3: ({ children }) => <h3 className={styles.headingTertiary}>{children}</h3>,
          h4: ({ children }) => <h4 className={styles.headingMinor}>{children}</h4>,
          input: ({ checked, disabled, type }) =>
            type === "checkbox" ? (
              <input
                aria-disabled={disabled || undefined}
                checked={checked}
                className={styles.checkbox}
                readOnly
                tabIndex={-1}
                type="checkbox"
              />
            ) : null,
          li: ({ children }) => <li className={styles.listItem}>{children}</li>,
          ol: ({ children }) => <ol className={styles.list}>{children}</ol>,
          p: ({ children }) => <p className={styles.paragraph}>{children}</p>,
          pre: ({ children }) => <pre className={styles.codeBlock}>{children}</pre>,
          table: ({ children }) => (
            <div className={styles.tableScroll}>
              <table className={styles.table}>{children}</table>
            </div>
          ),
          td: ({ children }) => <td className={styles.tableCell}>{children}</td>,
          th: ({ children }) => <th className={styles.tableHeader}>{children}</th>,
          ul: ({ children }) => <ul className={styles.list}>{children}</ul>,
        }}
        rehypePlugins={[rehypeSanitize]}
        remarkPlugins={[remarkGfm]}
      >
        {markdown}
      </ReactMarkdown>
    </div>
  );
}

function MarkdownCode({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const language = languageFromClassName(className);
  const code = nodeText(children);
  const highlightedLines = useShikiTokens(code, language);
  const classes = [
    styles.inlineCode,
    language ? styles.highlightedCode : undefined,
    className,
  ].filter(Boolean).join(" ");

  if (!language) {
    return <code className={classes}>{children}</code>;
  }

  return (
    <code className={classes}>
      {highlightedLines?.map((line, lineIndex) => (
        <span className={styles.codeLine} key={`line-${lineIndex}`}>
          {line.map((token, tokenIndex) => (
            <span key={`${lineIndex}-${tokenIndex}-${token.content}`} style={tokenStyle(token)}>
              {token.content}
            </span>
          ))}
        </span>
      )) ?? code}
    </code>
  );
}

function useShikiTokens(code: string, language: string | null): ShikiToken[][] | null {
  const [tokens, setTokens] = useState<ShikiToken[][] | null>(null);

  useEffect(() => {
    let isActive = true;
    setTokens(null);

    if (!language || code === "") {
      return () => {
        isActive = false;
      };
    }

    codeToTokens(code, {
      lang: shikiLanguage(language),
      theme: "github-light",
      tokenizeMaxLineLength: 500,
      tokenizeTimeLimit: 300,
    })
      .then((result) => {
        if (isActive) {
          setTokens(result.tokens);
        }
      })
      .catch(() => {
        if (isActive) {
          setTokens(null);
        }
      });

    return () => {
      isActive = false;
    };
  }, [code, language]);

  return tokens;
}

function languageFromClassName(className?: string): string | null {
  const match = className?.match(/(?:^|\s)language-([A-Za-z0-9_-]+)/);
  return match?.[1]?.toLowerCase() ?? null;
}

function nodeText(node: ReactNode): string {
  if (typeof node === "string" || typeof node === "number") {
    return String(node);
  }
  if (Array.isArray(node)) {
    return node.map(nodeText).join("");
  }
  return "";
}

function shikiLanguage(language: string): SupportedShikiLanguage | "text" {
  switch (language) {
    case "bash":
    case "diff":
    case "json":
    case "jsonc":
    case "shell":
    case "ts":
    case "tsx":
    case "typescript":
      return language;
    case "javascript":
    case "js":
      return "typescript";
    case "shellscript":
    case "sh":
    case "zsh":
      return "shell";
    case "patch":
      return "diff";
    default:
      return "text";
  }
}

function tokenStyle(token: ShikiToken): CSSProperties {
  const style: CSSProperties = {};
  if (token.color) {
    style.color = token.color;
  }
  if (token.fontStyle !== undefined && token.fontStyle > 0) {
    if ((token.fontStyle & 1) === 1) {
      style.fontStyle = "italic";
    }
    if ((token.fontStyle & 2) === 2) {
      style.fontWeight = 700;
    }
    if ((token.fontStyle & 4) === 4) {
      style.textDecorationLine = "underline";
    }
  }
  return style;
}

function rewriteAttachmentURLs(content: string): string {
  return content.replace(/attachment:\/\/([A-Za-z0-9_-]+)/g, (_, id: string) => {
    return `${issueTrackerURL()}/api/v1/attachments/${id}/content`;
  });
}

function issueTrackerURL(): string {
  return ("/tracker").replace(/\/$/, "");
}

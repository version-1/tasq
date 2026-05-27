import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Tasq",
  description: "Symphony-compatible task queue orchestrator",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}

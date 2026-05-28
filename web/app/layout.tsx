import type { Metadata } from "next";
import { Layout } from "@/components/layout";
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
      <body>
        <Layout>{children}</Layout>
      </body>
    </html>
  );
}

"use client";

import { Layout } from "@/components/layout";
import "./globals.css";

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

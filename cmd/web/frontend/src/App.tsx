import { Navigate, Route, Routes } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
import { IssueDetailLayout } from "./components/layout/issue";
import ConversationRoute from "./app/issues/[id]/conversations/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import SettingsPage from "./app/settings/page";
import { DefaultLayout } from "./components/layout/default";
import type { HeaderPageLink } from "./components/layout/header";
import { Layout } from "./components/layout";
import { ProjectLayout } from "./components/layout/project";

const projectHeaderPages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "settings", href: "/settings", titleKey: "header.settings" },
] satisfies readonly HeaderPageLink[];

const issueHeaderPages = [
  { key: "detail", href: "/issues/", titleKey: "issues.detailPage.detailTab" },
  {
    key: "conversations",
    href: "/issues/{id}/conversations",
    titleKey: "issues.detailPage.conversationTab",
  },
] satisfies readonly HeaderPageLink[];

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route
          path="/issues"
          element={
            <ProjectLayout pages={projectHeaderPages}>
              <IssuesPage />
            </ProjectLayout>
          }
        />
        <Route
          path="/projects/:projectKey/issues"
          element={
            <ProjectLayout pages={projectHeaderPages}>
              <IssuesPage />
            </ProjectLayout>
          }
        />
        <Route
          path="/issues/:id"
          element={
            <IssueDetailLayout pages={issueHeaderPages}>
              <IssueDetailRoute />
            </IssueDetailLayout>
          }
        />
        <Route
          path="/issues/:id/conversations"
          element={
            <IssueDetailLayout pages={issueHeaderPages}>
              <ConversationRoute />
            </IssueDetailLayout>
          }
        />
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={
            <IssueDetailLayout pages={issueHeaderPages}>
              <ConversationRoute />
            </IssueDetailLayout>
          }
        />
        <Route
          path="/dashboard"
          element={
            <DefaultLayout pages={[]}>
              <DashboardPage />
            </DefaultLayout>
          }
        />
        <Route
          path="/settings"
          element={
            <DefaultLayout pages={[]}>
              <SettingsPage />
            </DefaultLayout>
          }
        />
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}

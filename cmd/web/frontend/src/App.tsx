import { Navigate, Route, Routes } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
import { IssueDetailLayout } from "./components/layout/issue";
import ConversationRoute from "./app/issues/[id]/conversations/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import SettingsPage from "./app/settings/page";
import { DefaultLayout } from "./components/layout/default";
import { Layout } from "./components/layout";
import { ProjectLayout } from "./components/layout/project";
import { TabsProvider, type TabPageLink } from "./context/tabs";

const projectHeaderPages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "settings", href: "/settings", titleKey: "header.settings" },
] satisfies readonly TabPageLink[];

const issueHeaderPages = [
  { key: "detail", href: "/issues/:id", titleKey: "issues.detailPage.detailTab" },
  {
    key: "conversations",
    href: "/issues/:id/conversations",
    titleKey: "issues.detailPage.conversationTab",
  },
] satisfies readonly TabPageLink[];

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route
          path="/issues"
          element={
            <TabsProvider activeKey="issues" pages={projectHeaderPages}>
              <ProjectLayout>
                <IssuesPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/projects/:projectKey/issues"
          element={
            <TabsProvider activeKey="issues" pages={projectHeaderPages}>
              <ProjectLayout>
                <IssuesPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id"
          element={
            <TabsProvider activeKey="detail" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <IssueDetailRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id/conversations"
          element={
            <TabsProvider activeKey="conversations" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <ConversationRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={
            <TabsProvider activeKey="conversations" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <ConversationRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/dashboard"
          element={
            <TabsProvider>
              <DefaultLayout>
                <DashboardPage />
              </DefaultLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/settings"
          element={
            <TabsProvider>
              <DefaultLayout>
                <SettingsPage />
              </DefaultLayout>
            </TabsProvider>
          }
        />
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}

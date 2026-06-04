import { Navigate, Route, Routes } from "react-router-dom";
import AgentsPage from "./app/agents/page";
import DashboardPage from "./app/dashboard/page";
import IssueDetailRoute from "./app/issues/detail/page";
import IssuesPage from "./app/issues/page";
import SettingsPage from "./app/settings/page";
import { Layout } from "./components/layout";

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/issues" replace />} />
        <Route path="/issues" element={<IssuesPage />} />
        <Route path="/issues/detail" element={<IssueDetailRoute />} />
        <Route path="/agents" element={<AgentsPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}

import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import React from 'react';

// 懒加载页面组件
const AlertList = React.lazy(() => import('./pages/alerts/AlertList'));
const RuleList = React.lazy(() => import('./pages/alerts/RuleList'));
const KnowledgeList = React.lazy(() => import('./pages/knowledge/KnowledgeList'));
const KnowledgeDetail = React.lazy(() => import('./pages/knowledge/KnowledgeDetail'));
const TemplateList = React.lazy(() => import('./pages/notifications/TemplateList'));
const GroupList = React.lazy(() => import('./pages/notifications/GroupList'));
const Settings = React.lazy(() => import('./pages/settings/Settings'));

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <Router>
        <Routes>
          <Route path="/" element={<MainLayout />}>
            <Route index element={<Navigate to="/alerts" replace />} />
            <Route
              path="alerts"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <AlertList />
                </React.Suspense>
              }
            />
            <Route
              path="rules"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <RuleList />
                </React.Suspense>
              }
            />
            <Route
              path="knowledge"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <KnowledgeList />
                </React.Suspense>
              }
            />
            <Route
              path="knowledge/:id"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <KnowledgeDetail />
                </React.Suspense>
              }
            />
            <Route
              path="templates"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <TemplateList />
                </React.Suspense>
              }
            />
            <Route
              path="groups"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <GroupList />
                </React.Suspense>
              }
            />
            <Route
              path="settings"
              element={
                <React.Suspense fallback={<div>Loading...</div>}>
                  <Settings />
                </React.Suspense>
              }
            />
          </Route>
        </Routes>
      </Router>
    </ConfigProvider>
  );
}

export default App;

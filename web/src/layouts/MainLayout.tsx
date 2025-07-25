import React from 'react';
import { Layout, Menu } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
import {
  AlertOutlined,
  BellOutlined,
  SettingOutlined,
  BookOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';

const { Header, Content, Sider } = Layout;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();

  const menuItems = [
    {
      key: 'alerts',
      icon: <AlertOutlined />,
      label: '告警管理',
      children: [
        { key: '/alerts', label: '告警列表' },
        { key: '/rules', label: '告警规则' },
      ],
    },
    {
      key: 'notifications',
      icon: <BellOutlined />,
      label: '通知管理',
      children: [
        { key: '/templates', label: '通知模板' },
        { key: '/groups', label: '通知组' },
      ],
    },
    {
      key: 'knowledges',
      icon: <BookOutlined />,
      label: '知识库管理',
      children: [
        { key: '/knowledge', label: '知识库列表' },
      ],
    },
    {
      key: 'providers',
      icon: <DatabaseOutlined />,
      label: '数据源管理',
      children: [
        { key: '/providers', label: '数据源列表' },
      ],
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ padding: 0, background: '#fff' }}>
        <div style={{ float: 'left', width: 200, height: 31, margin: '16px 24px 16px 0', background: 'rgba(255, 255, 255, 0.2)' }}>
          Alert Agent
        </div>
      </Header>
      <Layout>
        <Sider width={200} style={{ background: '#fff' }}>
          <Menu
            mode="inline"
            defaultSelectedKeys={['alerts']}
            style={{ height: '100%', borderRight: 0 }}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
          />
        </Sider>
        <Layout style={{ padding: '24px' }}>
          <Content style={{ padding: 24, margin: 0, minHeight: 280, background: '#fff' }}>
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
import React, { useState } from 'react';
import { Layout, Menu, theme, Dropdown, Avatar, Space } from 'antd';
import type { MenuProps } from 'antd';
import {
  DesktopOutlined,
  PieChartOutlined,
  UserOutlined,
  TrophyOutlined,
  LogoutOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';

const { Header, Content, Footer, Sider } = Layout;

const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();
  
  const navigate = useNavigate();
  const location = useLocation();

  const items = [
    {
      key: '/',
      icon: <PieChartOutlined />,
      label: '仪表盘',
    },
    {
      key: '/challenges',
      icon: <DesktopOutlined />,
      label: '实战靶场',
    },
    {
      key: '/leaderboard',
      icon: <TrophyOutlined />,
      label: '排行榜',
    },
  ];

  const userMenu: MenuProps['items'] = [
    {
      key: 'profile',
      label: '个人中心',
      icon: <UserOutlined />,
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      label: '设置',
      icon: <SettingOutlined />,
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      label: '退出登录',
      icon: <LogoutOutlined />,
      onClick: () => navigate('/login'),
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={(value) => setCollapsed(value)}>
        <div style={{ height: 32, margin: 16, background: 'rgba(255, 255, 255, 0.2)', textAlign: 'center', color: '#fff', lineHeight: '32px', fontWeight: 'bold', overflow: 'hidden' }}>
          网络靶场
        </div>
        <Menu 
          theme="dark" 
          defaultSelectedKeys={['/']} 
          selectedKeys={[location.pathname]}
          mode="inline" 
          items={items} 
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: '0 24px', background: colorBgContainer, display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
          <Dropdown menu={{ items: userMenu }}>
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#87d068' }} />
              <span>Guest User</span>
            </Space>
          </Dropdown>
        </Header>
        <Content style={{ margin: '0 16px' }}>
          <div
            style={{
              padding: 24,
              minHeight: 360,
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
              marginTop: 16
            }}
          >
            <Outlet />
          </div>
        </Content>
        <Footer style={{ textAlign: 'center' }}>
          Cyber Range ©{new Date().getFullYear()} Created by Trae
        </Footer>
      </Layout>
    </Layout>
  );
};

export default MainLayout;

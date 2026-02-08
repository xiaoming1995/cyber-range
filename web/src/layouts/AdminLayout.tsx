import React, { useMemo, useState } from 'react';
import { Avatar, Dropdown, Layout, Menu, Space, Tooltip, theme } from 'antd';
import type { MenuProps } from 'antd';
import {
  DashboardOutlined,
  AppstoreOutlined,
  DeploymentUnitOutlined,
  FileSearchOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  CloudServerOutlined,
} from '@ant-design/icons';
import { Navigate, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { adminLogout, isAdminAuthed } from '../admin/auth';

const { Header, Content, Footer, Sider } = Layout;

const AdminLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = useMemo(
    () => [
      { key: '/admin/overview', icon: <DashboardOutlined />, label: '总览' },
      { key: '/admin/challenges', icon: <AppstoreOutlined />, label: '题库管理' },
      { key: '/admin/instances', icon: <DeploymentUnitOutlined />, label: '实例管理' },
      { key: '/admin/images', icon: <CloudServerOutlined />, label: '镜像管理' },
      { key: '/admin/submissions', icon: <FileSearchOutlined />, label: '提交记录' },
      { key: '/admin/logs', icon: <FileSearchOutlined />, label: 'API 日志' },
      {
        key: 'disabled:/admin/users',
        icon: <UserOutlined style={{ opacity: 0.5 }} />,
        label: (
          <Tooltip title="建设中">
            <span style={{ opacity: 0.5 }}>用户管理</span>
          </Tooltip>
        ),
      },
      {
        key: 'disabled:/admin/settings',
        icon: <SettingOutlined style={{ opacity: 0.5 }} />,
        label: (
          <Tooltip title="建设中">
            <span style={{ opacity: 0.5 }}>系统设置</span>
          </Tooltip>
        ),
      },
    ],
    [],
  );

  const userMenuItems: MenuProps['items'] = useMemo(
    () => [
      {
        key: 'logout',
        label: '退出登录',
        icon: <LogoutOutlined />,
        onClick: () => {
          adminLogout();
          navigate('/admin/login');
        },
      },
    ],
    [navigate],
  );

  const selectedKey = useMemo(() => {
    const pathname = location.pathname;
    if (pathname === '/admin') return '/admin/overview';
    if (pathname.startsWith('/admin/challenges/new')) return '/admin/challenges';
    if (pathname.startsWith('/admin/challenges')) return '/admin/challenges';
    if (pathname.startsWith('/admin/instances')) return '/admin/instances';
    if (pathname.startsWith('/admin/images')) return '/admin/images';
    if (pathname.startsWith('/admin/submissions')) return '/admin/submissions';
    if (pathname.startsWith('/admin/logs')) return '/admin/logs';
    if (pathname.startsWith('/admin/overview')) return '/admin/overview';
    return '';
  }, [location.pathname]);

  if (!isAdminAuthed()) return <Navigate to="/admin/login" replace />;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div
          style={{
            height: 32,
            margin: 16,
            background: 'rgba(255, 255, 255, 0.2)',
            textAlign: 'center',
            color: '#fff',
            lineHeight: '32px',
            fontWeight: 'bold',
            overflow: 'hidden',
          }}
        >
          后台管理
        </div>
        <Menu
          theme="dark"
          selectedKeys={[selectedKey]}
          mode="inline"
          items={menuItems}
          onClick={({ key }) => {
            if (String(key).startsWith('disabled:')) return;
            navigate(String(key));
          }}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
          }}
        >
          <Dropdown menu={{ items: userMenuItems }}>
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
              <span>Admin</span>
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
              marginTop: 16,
            }}
          >
            <Outlet />
          </div>
        </Content>
        <Footer style={{ textAlign: 'center' }}>
          Cyber Range Admin ©{new Date().getFullYear()}
        </Footer>
      </Layout>
    </Layout>
  );
};

export default AdminLayout;

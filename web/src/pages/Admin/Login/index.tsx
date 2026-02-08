import React, { useState } from 'react';
import { Alert, Button, Card, Form, Input, Typography, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { setAdminAuth } from '../../../admin/auth';
import { adminLogin } from '../../../api/admin';

const AdminLogin: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLogin = async (values: { username: string; password: string }) => {
    setLoading(true);
    setError('');

    try {
      const result = await adminLogin(values.username, values.password);
      setAdminAuth(result.token, result.admin);
      message.success('登录成功');
      navigate('/admin/overview', { replace: true });
    } catch (err: any) {
      setError(err.message || '登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}>
      <Card style={{ width: 420 }}>
        <Typography.Title level={3} style={{ marginTop: 0 }}>
          管理员登录
        </Typography.Title>

        {error && (
          <Alert
            message={error}
            type="error"
            closable
            onClose={() => setError('')}
            style={{ marginBottom: 16 }}
          />
        )}

        <Form layout="vertical" onFinish={handleLogin}>
          <Form.Item name="username" label="用户名/邮箱" rules={[{ required: true, message: '请输入用户名/邮箱' }]}>
            <Input autoComplete="username" placeholder="admin" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password autoComplete="current-password" placeholder="admin123" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>
            登录
          </Button>
        </Form>

        <div style={{ marginTop: 16, textAlign: 'center', color: '#999', fontSize: 12 }}>
          默认账号：admin / admin123
        </div>
      </Card>
    </div>
  );
};

export default AdminLogin;

import React from 'react';
import { Card, Button, Checkbox, Form, Input, Typography, Tabs } from 'antd';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title } = Typography;

const Login: React.FC = () => {
  const navigate = useNavigate();

  const onFinish = (values: Record<string, unknown>) => {
    console.log('Received values of form: ', values);
    navigate('/');
  };

  return (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: '#f0f2f5'
    }}>
      <Card style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <Title level={3}>Cyber Range</Title>
          <p>网络安全实战演练平台</p>
        </div>
        
        <Tabs
          defaultActiveKey="1"
          centered
          items={[
            {
              key: '1',
              label: '账号登录',
              children: (
                <Form
                  name="normal_login"
                  initialValues={{ remember: true }}
                  onFinish={onFinish}
                >
                  <Form.Item
                    name="username"
                    rules={[{ required: true, message: '请输入用户名!' }]}
                  >
                    <Input prefix={<UserOutlined />} placeholder="用户名" size="large" />
                  </Form.Item>
                  <Form.Item
                    name="password"
                    rules={[{ required: true, message: '请输入密码!' }]}
                  >
                    <Input
                      prefix={<LockOutlined />}
                      type="password"
                      placeholder="密码"
                      size="large"
                    />
                  </Form.Item>
                  <Form.Item>
                    <Form.Item name="remember" valuePropName="checked" noStyle>
                      <Checkbox>自动登录</Checkbox>
                    </Form.Item>
                    <a style={{ float: 'right' }} href="">
                      忘记密码
                    </a>
                  </Form.Item>

                  <Form.Item>
                    <Button type="primary" htmlType="submit" style={{ width: '100%' }} size="large">
                      登录
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: '2',
              label: '注册新账号',
              children: (
                <div style={{ textAlign: 'center', padding: '20px 0' }}>
                  <p>注册功能暂未开放</p>
                </div>
              )
            }
          ]}
        />
      </Card>
    </div>
  );
};

export default Login;

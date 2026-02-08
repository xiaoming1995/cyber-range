import React from 'react';
import { Card, Row, Col, Avatar, Typography, List, Divider } from 'antd';
import { UserOutlined, MailOutlined, EnvironmentOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { Radar } from '@ant-design/plots';

const { Title, Text } = Typography;

const Profile: React.FC = () => {
  // Mock data for skill radar
  const data = [
    { item: 'Web', score: 70 },
    { item: 'Pwn', score: 40 },
    { item: 'Crypto', score: 50 },
    { item: 'Reverse', score: 30 },
    { item: 'Misc', score: 80 },
    { item: 'Mobile', score: 20 },
  ];

  const config = {
    data: data.map((d) => ({ ...d, user: 'User' })),
    xField: 'item',
    yField: 'score',
    meta: {
      score: {
        alias: '能力值',
        min: 0,
        max: 100,
      },
    },
    xAxis: {
      line: null,
      tickLine: null,
      grid: {
        line: {
          style: {
            lineDash: null,
          },
        },
      },
    },
    area: {},
    // point: {
    //   size: 2,
    // },
  };

  const activities = [
    { title: '完成挑战 "Nginx 基础挑战"', time: '2小时前' },
    { title: '获得徽章 "Web 新手"', time: '1天前' },
    { title: '注册账号', time: '3天前' },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Row gutter={24}>
        <Col span={8}>
          <Card bordered={false} style={{ textAlign: 'center' }}>
            <Avatar size={100} icon={<UserOutlined />} style={{ marginBottom: 16 }} />
            <Title level={3}>Guest User</Title>
            <Text type="secondary">网络安全爱好者</Text>
            <Divider />
            <div style={{ textAlign: 'left' }}>
              <p><MailOutlined style={{ marginRight: 8 }} /> guest@example.com</p>
              <p><EnvironmentOutlined style={{ marginRight: 8 }} /> 中国, 北京</p>
              <p><SafetyCertificateOutlined style={{ marginRight: 8 }} /> 积分: 100</p>
            </div>
          </Card>
        </Col>
        <Col span={16}>
          <Card title="技能雷达图" bordered={false} style={{ marginBottom: 24 }}>
            <Radar {...config} style={{ height: 300 }} />
          </Card>
          <Card title="最近动态" bordered={false}>
            <List
              dataSource={activities}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={item.title}
                    description={item.time}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Profile;

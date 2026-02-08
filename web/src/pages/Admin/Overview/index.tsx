import React, { useEffect, useState } from 'react';
import { Card, Col, List, Row, Statistic, Table, Tag, Typography, message } from 'antd';
import { getOverviewStats } from '../../../api/admin';
import type { OverviewStats } from '../../../api/admin';

const AdminOverview: React.FC = () => {
  const [stats, setStats] = useState<OverviewStats>({
    todayInstances: 0,
    runningInstances: 0,
    todaySubmissions: 0,
    todayCorrectRate: 0,
    recentSubmissions: [],
    hotChallenges: [],
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        const data = await getOverviewStats();
        setStats(data);
      } catch (error) {
        message.error('获取总览数据失败');
      }
    };
    fetchData();
  }, []);

  return (
    <div>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        总览
      </Typography.Title>

      <Row gutter={16}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic title="今日启动实例数" value={stats.todayInstances} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic title="运行中实例数" value={stats.runningInstances} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic title="今日提交数" value={stats.todaySubmissions} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic title="今日正确率" value={stats.todayCorrectRate} suffix="%" />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col xs={24} md={16}>
          <Card title="最近提交" bodyStyle={{ padding: 0 }}>
            <Table
              size="small"
              pagination={false}
              dataSource={(stats.recentSubmissions || []).map((s) => ({ ...s, key: s.id }))}
              columns={[
                {
                  title: '时间',
                  dataIndex: 'createdAt',
                  width: 200,
                  render: (v: string) => new Date(v).toLocaleString(),
                },
                { title: '用户', dataIndex: 'userDisplayName', width: 120 },
                { title: '题目', dataIndex: 'challengeTitle' },
                {
                  title: '结果',
                  dataIndex: 'result',
                  width: 90,
                  render: (v: string) => (v === 'correct' ? <Tag color="green">正确</Tag> : <Tag color="red">错误</Tag>),
                },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title="题目热度 Top5">
            <List
              dataSource={stats.hotChallenges || []}
              renderItem={(item, idx) => (
                <List.Item>
                  <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                    <span>
                      {idx + 1}. {item.title}
                    </span>
                    <span style={{ opacity: 0.75 }}>{item.count}</span>
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default AdminOverview;

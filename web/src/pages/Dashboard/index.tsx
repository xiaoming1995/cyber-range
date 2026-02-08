import React from 'react';
import { Card, Col, Row, Statistic } from 'antd';
import { ArrowUpOutlined, RocketOutlined } from '@ant-design/icons';

const Dashboard: React.FC = () => (
  <div>
    <h2>仪表盘</h2>
    <Row gutter={16}>
      <Col span={8}>
        <Card>
          <Statistic
            title="活跃实例"
            value={3}
            precision={0}
            valueStyle={{ color: '#3f8600' }}
            prefix={<ArrowUpOutlined />}
            suffix=""
          />
        </Card>
      </Col>
      <Col span={8}>
        <Card>
          <Statistic
            title="总挑战数"
            value={12}
            precision={0}
            valueStyle={{ color: '#cf1322' }}
            prefix={<RocketOutlined />}
            suffix=""
          />
        </Card>
      </Col>
    </Row>
  </div>
);

export default Dashboard;

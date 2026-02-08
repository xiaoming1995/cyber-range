import React from 'react';
import { Table, Avatar, Tag, Card } from 'antd';
import { UserOutlined, TrophyOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

interface RankUser {
  rank: number;
  name: string;
  points: number;
  badges: string[];
  solved: number;
}

const data: RankUser[] = [
  { rank: 1, name: 'CyberMaster', points: 1500, badges: ['PwnGod', 'WebMaster'], solved: 42 },
  { rank: 2, name: 'Alice', points: 1350, badges: ['FastSolver'], solved: 38 },
  { rank: 3, name: 'Bob', points: 1200, badges: [], solved: 35 },
  { rank: 4, name: 'Dave', points: 900, badges: ['Rookie'], solved: 20 },
  { rank: 5, name: 'Eve', points: 850, badges: [], solved: 18 },
];

const Leaderboard: React.FC = () => {
  const columns: ColumnsType<RankUser> = [
    {
      title: '排名',
      dataIndex: 'rank',
      key: 'rank',
      render: (rank) => {
        let color = '#8c8c8c';
        if (rank === 1) color = '#fadb14';
        if (rank === 2) color = '#d4b106';
        if (rank === 3) color = '#cd7f32';
        return <span style={{ fontWeight: 'bold', color, fontSize: 16 }}>#{rank}</span>;
      },
    },
    {
      title: '用户',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <span>
          <Avatar icon={<UserOutlined />} style={{ marginRight: 8 }} />
          {text}
        </span>
      ),
    },
    {
      title: '徽章',
      dataIndex: 'badges',
      key: 'badges',
      render: (badges: string[]) => (
        <>
          {badges.map((badge) => (
            <Tag color="blue" key={badge}>
              {badge}
            </Tag>
          ))}
        </>
      ),
    },
    {
      title: '解题数',
      dataIndex: 'solved',
      key: 'solved',
    },
    {
      title: '积分',
      dataIndex: 'points',
      key: 'points',
      render: (points) => (
        <span style={{ color: '#cf1322', fontWeight: 'bold' }}>
          {points} <TrophyOutlined />
        </span>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card title="积分排行榜" bordered={false}>
        <Table columns={columns} dataSource={data} rowKey="rank" pagination={false} />
      </Card>
    </div>
  );
};

export default Leaderboard;

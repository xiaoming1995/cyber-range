import React, { useEffect, useState } from 'react';
import { Alert, Button, Drawer, Form, Input, Select, Space, Table, Tag, Typography } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { listSubmissions } from '../../../api/admin';
import type { SelectProps } from 'antd';

const resultOptions: NonNullable<SelectProps['options']> = [
  { label: '正确', value: 'correct' },
  { label: '错误', value: 'wrong' },
];

interface Submission {
  id: string;
  user_id: string;
  challenge_id: string;
  challenge_title?: string;
  user_display_name?: string;
  flag: string;
  is_correct: boolean;
  points: number;
  submitted_at: string;
}

const AdminSubmissions: React.FC = () => {
  const [filters] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<{ list: Submission[]; total: number }>({ list: [], total: 0 });
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });
  const [selected, setSelected] = useState<Submission | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const fetchSubmissions = async (page = 1, pageSize = 10, result?: string) => {
    setLoading(true);
    setError(null);
    try {
      const res = await listSubmissions({
        page,
        pageSize,
        result: result || undefined,
      });
      setData({ list: res.list || [], total: res.total || 0 });
      setPagination({ current: page, pageSize });
    } catch (err: any) {
      setError(err.message || '获取提交记录失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSubmissions();
  }, []);

  const handleFilterChange = () => {
    const { result } = filters.getFieldsValue();
    fetchSubmissions(1, pagination.pageSize, result);
  };

  const handleTableChange = (paginationInfo: any) => {
    const { result } = filters.getFieldsValue();
    fetchSubmissions(paginationInfo.current, paginationInfo.pageSize, result);
  };

  return (
    <div>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        提交记录
      </Typography.Title>

      <Form form={filters} layout="inline" style={{ margin: '16px 0' }} onValuesChange={handleFilterChange}>
        <Form.Item name="keyword">
          <Input placeholder="搜索用户或题目" allowClear style={{ width: 260 }} />
        </Form.Item>
        <Form.Item name="result">
          <Select placeholder="结果" allowClear style={{ width: 140 }} options={resultOptions} />
        </Form.Item>
        <Form.Item>
          <Button
            onClick={() => {
              filters.resetFields();
              fetchSubmissions();
            }}
          >
            重置
          </Button>
        </Form.Item>
        <Form.Item>
          <Button type="primary" icon={<ReloadOutlined />} onClick={() => fetchSubmissions(pagination.current, pagination.pageSize)}>
            刷新
          </Button>
        </Form.Item>
      </Form>

      {error && <Alert message={error} type="error" style={{ marginBottom: 16 }} />}

      <Table
        rowKey="id"
        dataSource={data.list}
        loading={loading}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: data.total,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={handleTableChange}
        columns={[
          {
            title: '时间',
            dataIndex: 'submitted_at',
            width: 200,
            render: (v: string) => v ? new Date(v).toLocaleString() : '-',
          },
          { title: '用户', dataIndex: 'user_display_name', width: 140, render: (v: string) => v || '-' },
          { title: '题目', dataIndex: 'challenge_title', render: (v: string) => v || '-' },
          {
            title: '结果',
            dataIndex: 'is_correct',
            width: 110,
            render: (v: boolean) => (v ? <Tag color="green">正确</Tag> : <Tag color="red">错误</Tag>),
          },
          {
            title: '积分',
            dataIndex: 'points',
            width: 80,
          },
          {
            title: '操作',
            key: 'actions',
            width: 100,
            render: (_: unknown, record: Submission) => (
              <Button
                size="small"
                onClick={() => {
                  setSelected(record);
                  setDrawerOpen(true);
                }}
              >
                查看
              </Button>
            ),
          },
        ]}
      />

      <Drawer
        open={drawerOpen}
        onClose={() => {
          setDrawerOpen(false);
          setSelected(null);
        }}
        title="提交详情"
        width={680}
      >
        {!selected ? null : (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <div>
              <Typography.Text strong>时间：</Typography.Text> {selected.submitted_at ? new Date(selected.submitted_at).toLocaleString() : '-'}
            </div>
            <div>
              <Typography.Text strong>用户：</Typography.Text> {selected.user_display_name || selected.user_id || '-'}
            </div>
            <div>
              <Typography.Text strong>题目：</Typography.Text> {selected.challenge_title || selected.challenge_id || '-'}
            </div>
            <div>
              <Typography.Text strong>结果：</Typography.Text>{' '}
              {selected.is_correct ? <Tag color="green">正确</Tag> : <Tag color="red">错误</Tag>}
            </div>
            <div>
              <Typography.Text strong>积分：</Typography.Text> {selected.points}
            </div>
            <div>
              <Typography.Text type="secondary">默认不展示 Flag 明文，如需查看应在后端权限控制后开放。</Typography.Text>
            </div>
          </Space>
        )}
      </Drawer>
    </div>
  );
};

export default AdminSubmissions;

import React, { useMemo, useState } from 'react';
import { Button, Drawer, Form, Input, Select, Space, Table, Tag, Typography } from 'antd';
import type { AdminSubmission, SubmissionResult } from '../../../admin/types';
import { listSubmissions } from '../../../admin/store';
import type { SelectProps } from 'antd';

const resultOptions: NonNullable<SelectProps['options']> = [
  { label: '正确', value: 'correct' },
  { label: '错误', value: 'wrong' },
];

const AdminSubmissions: React.FC = () => {
  const [filters] = Form.useForm();
  const [refreshKey, setRefreshKey] = useState(0);
  const [selected, setSelected] = useState<AdminSubmission | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const submissions = useMemo(() => listSubmissions(), [refreshKey]);

  const filtered = useMemo(() => {
    const { keyword, result } = filters.getFieldsValue();
    const kw = String(keyword || '').trim().toLowerCase();
    return submissions.filter((s) => {
      if (kw && !(s.userDisplayName.toLowerCase().includes(kw) || s.challengeTitle.toLowerCase().includes(kw))) return false;
      if (result && s.result !== result) return false;
      return true;
    });
  }, [submissions, filters, refreshKey]);

  return (
    <div>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        提交记录
      </Typography.Title>

      <Form form={filters} layout="inline" style={{ margin: '16px 0' }} onValuesChange={() => setRefreshKey((k) => k + 1)}>
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
              setRefreshKey((k) => k + 1);
            }}
          >
            重置
          </Button>
        </Form.Item>
      </Form>

      <Table
        rowKey="id"
        dataSource={filtered.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())}
        pagination={{ pageSize: 10 }}
        columns={[
          {
            title: '时间',
            dataIndex: 'createdAt',
            width: 200,
            render: (v: string) => new Date(v).toLocaleString(),
          },
          { title: '用户', dataIndex: 'userDisplayName', width: 140 },
          { title: '题目', dataIndex: 'challengeTitle' },
          {
            title: '结果',
            dataIndex: 'result',
            width: 110,
            render: (v: SubmissionResult) => (v === 'correct' ? <Tag color="green">正确</Tag> : <Tag color="red">错误</Tag>),
          },
          { title: 'IP', dataIndex: 'ip', width: 160, render: (v: string | undefined) => v || '-' },
          {
            title: '操作',
            key: 'actions',
            width: 100,
            render: (_: unknown, record: AdminSubmission) => (
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
              <Typography.Text strong>时间：</Typography.Text> {new Date(selected.createdAt).toLocaleString()}
            </div>
            <div>
              <Typography.Text strong>用户：</Typography.Text> {selected.userDisplayName}
            </div>
            <div>
              <Typography.Text strong>题目：</Typography.Text> {selected.challengeTitle}
            </div>
            <div>
              <Typography.Text strong>结果：</Typography.Text>{' '}
              {selected.result === 'correct' ? <Tag color="green">正确</Tag> : <Tag color="red">错误</Tag>}
            </div>
            <div>
              <Typography.Text strong>IP：</Typography.Text> {selected.ip || '-'}
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

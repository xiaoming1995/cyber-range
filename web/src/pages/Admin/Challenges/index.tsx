import React, { useMemo, useState } from 'react';
import { Button, Checkbox, Divider, Dropdown, Form, Input, Modal, Popover, Select, Space, Table, Tag, Typography, message, Spin, Alert } from 'antd';
import { DownOutlined, PlusOutlined, SettingOutlined, ReloadOutlined } from '@ant-design/icons';
import * as adminAPI from '../../../api/admin';
import { useChallenges } from '../../../admin/hooks';
import ChallengeEditDrawer from './ChallengeEditDrawer';
import { useNavigate } from 'react-router-dom';
import type { SelectProps, TableColumnsType } from 'antd';

type Challenge = adminAPI.Challenge;
type ChallengeStatus = Challenge['status'];

type ColumnKey = 'image' | 'port' | 'createdAt';

const COLUMN_STORAGE_KEY = 'cyber_range_admin_challenge_columns_v1';

const optionalColumns: { key: ColumnKey; label: string }[] = [
  { key: 'image', label: '镜像' },
  { key: 'port', label: '端口' },
  { key: 'createdAt', label: '创建时间' },
];

function loadColumnKeys(): ColumnKey[] {
  const raw = localStorage.getItem(COLUMN_STORAGE_KEY);
  if (!raw) return [];
  try {
    const arr = JSON.parse(raw) as ColumnKey[];
    return Array.isArray(arr) ? arr : [];
  } catch {
    return [];
  }
}

function saveColumnKeys(keys: ColumnKey[]) {
  localStorage.setItem(COLUMN_STORAGE_KEY, JSON.stringify(keys));
}

const categoryOptions: NonNullable<SelectProps['options']> = [
  { label: 'Web', value: 'Web' },
  { label: 'Pwn', value: 'Pwn' },
  { label: 'Crypto', value: 'Crypto' },
  { label: 'Reverse', value: 'Reverse' },
  { label: 'Misc', value: 'Misc' },
];

const difficultyOptions: NonNullable<SelectProps['options']> = [
  { label: 'Easy', value: 'Easy' },
  { label: 'Medium', value: 'Medium' },
  { label: 'Hard', value: 'Hard' },
];

const statusOptions: NonNullable<SelectProps['options']> = [
  { label: '上架', value: 'published' },
  { label: '下架', value: 'unpublished' },
];

const AdminChallenges: React.FC = () => {
  const [filters] = Form.useForm();
  const [selected, setSelected] = useState<Challenge | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [columnKeys, setColumnKeys] = useState<ColumnKey[]>(() => loadColumnKeys());
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const navigate = useNavigate();

  // 获取筛选参数
  const { keyword, category, difficulty, status } = Form.useWatch([], filters) || {};

  // 筛选条件变化时重置页码
  React.useEffect(() => {
    if (page !== 1) {
      setPage(1);
    }
  }, [category, difficulty, status, keyword]);

  // 使用真实 API
  const { challenges, total, loading, error, refetch } = useChallenges({
    page,
    pageSize,
    category,
    difficulty,
    status,
    search: keyword,
  });

  const openEditor = (challenge: Challenge) => {
    setSelected(challenge);
    setDrawerOpen(true);
  };

  // 更新状态
  const handleUpdateStatus = async (id: string, newStatus: ChallengeStatus) => {
    try {
      await adminAPI.updateChallengeStatus(id, newStatus);
      message.success('已更新');
      refetch();
    } catch (err: any) {
      message.error(err.message || '更新失败');
    }
  };

  // 删除题目
  const handleDelete = async (id: string) => {
    try {
      await adminAPI.deleteChallenge(id);
      message.success('已删除');
      refetch();
    } catch (err: any) {
      message.error(err.message || '删除失败');
    }
  };

  const columnSettings = (
    <div style={{ width: 220 }}>
      <Typography.Text strong>列设置</Typography.Text>
      <Divider style={{ margin: '8px 0' }} />
      <Checkbox.Group
        value={columnKeys}
        options={optionalColumns.map((c) => ({ label: c.label, value: c.key }))}
        onChange={(vals) => {
          const keys = vals as ColumnKey[];
          setColumnKeys(keys);
          saveColumnKeys(keys);
        }}
      />
      <Divider style={{ margin: '8px 0' }} />
      <Button
        size="small"
        onClick={() => {
          setColumnKeys([]);
          saveColumnKeys([]);
        }}
      >
        恢复默认
      </Button>
    </div>
  );

  const columns = useMemo<TableColumnsType<Challenge>>(() => {
    const base: TableColumnsType<Challenge> = [
      {
        title: '标题',
        dataIndex: 'title',
        render: (_: string, record: Challenge) => (
          <Typography.Link onClick={() => openEditor(record)}>{record.title}</Typography.Link>
        ),
      },
      { title: '分类', dataIndex: 'category', width: 110 },
      { title: '难度', dataIndex: 'difficulty', width: 110 },
      { title: '分值', dataIndex: 'points', width: 90 },
      {
        title: '状态',
        dataIndex: 'status',
        width: 140,
        render: (v: ChallengeStatus, record: Challenge) => (
          <Space>
            {v === 'published' ? <Tag color="green">上架</Tag> : <Tag>下架</Tag>}
            <Button
              size="small"
              onClick={() => {
                const next: ChallengeStatus = record.status === 'published' ? 'unpublished' : 'published';
                Modal.confirm({
                  title: next === 'published' ? '确认上架？' : '确认下架？',
                  content: next === 'published' ? '上架后用户侧可见。' : '下架后用户侧不可见。',
                  okText: '确认',
                  cancelText: '取消',
                  onOk: () => handleUpdateStatus(record.id, next),
                });
              }}
            >
              切换
            </Button>
          </Space>
        ),
      },
      {
        title: '更新时间',
        dataIndex: 'updated_at',
        width: 200,
        render: (v: string) => new Date(v).toLocaleString(),
      },
      {
        title: '操作',
        key: 'actions',
        width: 250,
        render: (_: unknown, record: Challenge) => {
          const next: ChallengeStatus = record.status === 'published' ? 'unpublished' : 'published';
          return (
            <Space>
              <Button size="small" onClick={() => openEditor(record)}>
                编辑
              </Button>
              <Button
                size="small"
                onClick={() => {
                  Modal.confirm({
                    title: next === 'published' ? '确认上架？' : '确认下架？',
                    okText: '确认',
                    cancelText: '取消',
                    onOk: () => handleUpdateStatus(record.id, next),
                  });
                }}
              >
                {next === 'published' ? '上架' : '下架'}
              </Button>
              <Button
                size="small"
                danger
                onClick={() => {
                  Modal.confirm({
                    title: '确认删除？',
                    content: '删除后不可恢复。',
                    okText: '删除',
                    cancelText: '取消',
                    okButtonProps: { danger: true },
                    onOk: () => handleDelete(record.id),
                  });
                }}
              >
                删除
              </Button>
              <Dropdown
                menu={{
                  items: [
                    {
                      key: 'copy',
                      label: '复制挑战',
                      onClick: () => {
                        navigator.clipboard.writeText(JSON.stringify(record, null, 2));
                        message.success('已复制到剪贴板');
                      },
                    },
                    { type: 'divider' },
                    {
                      key: 'toggle',
                      label: next === 'published' ? '上架' : '下架',
                      onClick: () => handleUpdateStatus(record.id, next),
                    },
                  ],
                }}
              >
                <Button size="small">
                  更多 <DownOutlined />
                </Button>
              </Dropdown>
            </Space>
          );
        },
      },
    ];

    const extras: TableColumnsType<Challenge> = [];
    if (columnKeys.includes('image')) {
      extras.push({ title: '镜像', dataIndex: 'image', width: 200 });
    }
    if (columnKeys.includes('port')) {
      extras.push({ title: '端口', dataIndex: 'port', width: 90 });
    }
    if (columnKeys.includes('createdAt')) {
      extras.push({
        title: '创建时间',
        dataIndex: 'created_at',
        width: 200,
        render: (v: string) => new Date(v).toLocaleString(),
      });
    }

    const idx = base.findIndex((c) => 'dataIndex' in c && c.dataIndex === 'updated_at');
    if (idx >= 0) return [...base.slice(0, idx), ...extras, ...base.slice(idx)];
    return [...base, ...extras];
  }, [columnKeys]);

  if (error) {
    return (
      <div>
        <Alert
          message="加载失败"
          description={error}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={refetch}>
              重试
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography.Title level={3} style={{ marginTop: 0 }}>
          题库管理
        </Typography.Title>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={refetch} loading={loading}>
            刷新
          </Button>
          <Popover content={columnSettings} trigger="click" placement="bottomRight">
            <Button icon={<SettingOutlined />}>列设置</Button>
          </Popover>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/admin/challenges/new')}>
            新建挑战
          </Button>
        </Space>
      </div>

      <Form
        form={filters}
        layout="inline"
        style={{ margin: '16px 0' }}
        initialValues={{ keyword: '', category: undefined, difficulty: undefined, status: undefined }}
      >
        <Form.Item name="keyword">
          <Input placeholder="搜索标题" allowClear style={{ width: 260 }} />
        </Form.Item>
        <Form.Item name="category">
          <Select
            placeholder="分类"
            allowClear
            style={{ width: 140 }}
            options={categoryOptions}
          />
        </Form.Item>
        <Form.Item name="difficulty">
          <Select placeholder="难度" allowClear style={{ width: 140 }} options={difficultyOptions} />
        </Form.Item>
        <Form.Item name="status">
          <Select placeholder="状态" allowClear style={{ width: 140 }} options={statusOptions} />
        </Form.Item>
        <Form.Item>
          <Button
            onClick={() => {
              filters.resetFields();
              refetch();
            }}
          >
            重置
          </Button>
        </Form.Item>
      </Form>

      <Spin spinning={loading}>
        <Table
          rowKey="id"
          dataSource={challenges}
          columns={columns}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => {
              setPage(page);
              setPageSize(pageSize);
            },
          }}
        />
      </Spin>

      {selected && (
        <ChallengeEditDrawer
          key={selected.id}
          open={drawerOpen}
          challenge={{
            id: selected.id,
            title: selected.title,
            category: selected.category as any, // API 返回的是 string，需要类型断言
            difficulty: selected.difficulty as any,
            points: selected.points,
            image: selected.image,
            image_id: (selected as any).image_id, // 传递 image_id
            docker_host_id: (selected as any).docker_host_id, // 传递 docker_host_id
            port: selected.port,
            memory_limit: (selected as any).memory_limit, // 传递资源限制
            cpu_limit: (selected as any).cpu_limit,
            status: selected.status,
            descriptionHtml: selected.description,
            hintHtml: selected.hint || '',
            flag: selected.flag, // 允许回显 flag
            createdAt: selected.created_at,
            updatedAt: selected.updated_at,
          }}
          onClose={() => {
            setDrawerOpen(false);
            setSelected(null);
          }}
          onSave={async (id: string, data: any) => {
            try {
              await adminAPI.updateChallenge(id, data);
              refetch();
              setDrawerOpen(false);
              setSelected(null);
            } catch (err: any) {
              message.error(err.message || '保存失败');
              throw err; // 继续抛出错误让 Drawer 捕获
            }
          }}
        />
      )}
    </div>
  );
};

export default AdminChallenges;

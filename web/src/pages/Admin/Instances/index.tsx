import React, { useEffect, useState } from 'react';
import { Button, Drawer, Form, Select, Space, Table, Tag, Typography, Progress, Spin, Alert, Input } from 'antd';
import { LoadingOutlined, ReloadOutlined } from '@ant-design/icons';
import { listInstances as listInstancesAPI, getInstanceStats, getInstanceLogs, type ContainerStats } from '../../../api/admin';
import type { SelectProps } from 'antd';

const statusOptions: NonNullable<SelectProps['options']> = [
  { label: '全部', value: '' },
  { label: '运行中', value: 'running' },
  { label: '已停止', value: 'stopped' },
  { label: '已过期', value: 'expired' },
];

// 格式化字节数
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 实例类型（来自后端）
interface Instance {
  id: string;
  user_id: string;
  challenge_id: string;
  challenge_title?: string;
  container_id: string;
  docker_host_id: string;
  port: number;
  status: 'running' | 'stopped' | 'expired';
  expires_at: string;
  created_at: string;
}

// 资源监控组件
const InstanceStatsPanel: React.FC<{ instanceId: string; status: string }> = ({ instanceId, status }) => {
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<ContainerStats | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = async () => {
    if (status !== 'running') return;
    setLoading(true);
    setError(null);
    try {
      const data = await getInstanceStats(instanceId);
      setStats(data);
    } catch (err: any) {
      setError(err.message || '获取统计失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, [instanceId, status]);

  if (status !== 'running') {
    return (
      <div style={{ padding: 16, color: '#999' }}>
        实例未运行，无法获取资源统计
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ padding: 16, textAlign: 'center' }}>
        <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
        <div style={{ marginTop: 8 }}>正在获取资源统计...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: 16 }}>
        <Typography.Text type="danger">{error}</Typography.Text>
        <Button icon={<ReloadOutlined />} onClick={fetchStats} style={{ marginLeft: 8 }} size="small">
          重试
        </Button>
      </div>
    );
  }

  if (!stats) return null;

  return (
    <div style={{ padding: '12px 16px', background: '#fafafa' }}>
      <Space size="large" wrap>
        <div>
          <Typography.Text type="secondary">CPU:</Typography.Text>{' '}
          <Typography.Text strong>{stats.cpu_percent.toFixed(1)}%</Typography.Text>
        </div>
        <div style={{ minWidth: 200 }}>
          <Typography.Text type="secondary">内存:</Typography.Text>{' '}
          <Progress
            percent={stats.memory_percent}
            size="small"
            format={() => `${formatBytes(stats.memory_usage)} / ${formatBytes(stats.memory_limit)}`}
            style={{ width: 180, display: 'inline-flex' }}
          />
        </div>
        <div>
          <Typography.Text type="secondary">网络:</Typography.Text>{' '}
          ↓{formatBytes(stats.network_rx)} ↑{formatBytes(stats.network_tx)}
        </div>
        <Button icon={<ReloadOutlined />} onClick={fetchStats} size="small">
          刷新
        </Button>
      </Space>
    </div>
  );
};

// 日志查看组件
const InstanceLogsPanel: React.FC<{ instanceId: string; containerId: string }> = ({ instanceId, containerId }) => {
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<string>('');
  const [error, setError] = useState<string | null>(null);
  const [tail, setTail] = useState(200);

  const fetchLogs = async () => {
    if (!containerId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await getInstanceLogs(instanceId, tail);
      setLogs(data.logs);
    } catch (err: any) {
      setError(err.message || '获取日志失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [instanceId, tail]);

  if (!containerId) {
    return (
      <div style={{ marginTop: 16 }}>
        <Typography.Title level={5}>容器日志</Typography.Title>
        <div style={{ color: '#999' }}>容器ID未知，无法获取日志</div>
      </div>
    );
  }

  return (
    <div style={{ marginTop: 16 }}>
      <Space style={{ marginBottom: 8 }}>
        <Typography.Title level={5} style={{ margin: 0 }}>容器日志</Typography.Title>
        <Select
          value={tail}
          onChange={setTail}
          size="small"
          style={{ width: 100 }}
          options={[
            { label: '100行', value: 100 },
            { label: '200行', value: 200 },
            { label: '500行', value: 500 },
            { label: '1000行', value: 1000 },
          ]}
        />
        <Button icon={<ReloadOutlined />} onClick={fetchLogs} size="small" loading={loading}>
          刷新
        </Button>
      </Space>

      {error && <Alert message={error} type="error" style={{ marginBottom: 8 }} />}

      {loading && !logs ? (
        <div style={{ textAlign: 'center', padding: 16 }}>
          <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
        </div>
      ) : (
        <Input.TextArea
          value={logs || '(暂无日志)'}
          readOnly
          rows={12}
          style={{
            fontFamily: 'monospace',
            fontSize: 12,
            background: '#1e1e1e',
            color: '#d4d4d4',
          }}
        />
      )}
    </div>
  );
};

const AdminInstances: React.FC = () => {
  const [filters] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<{ list: Instance[]; total: number }>({ list: [], total: 0 });
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });
  const [selected, setSelected] = useState<Instance | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [expandedRowKeys, setExpandedRowKeys] = useState<string[]>([]);

  const fetchInstances = async (page = 1, pageSize = 10, status?: string) => {
    setLoading(true);
    setError(null);
    try {
      const result = await listInstancesAPI({
        page,
        pageSize,
        status: status || undefined,
      });
      setData({ list: result.list || [], total: result.total || 0 });
      setPagination({ current: page, pageSize });
    } catch (err: any) {
      setError(err.message || '获取实例列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchInstances();
  }, []);

  const handleTableChange = (paginationInfo: any) => {
    const { status } = filters.getFieldsValue();
    fetchInstances(paginationInfo.current, paginationInfo.pageSize, status);
  };

  const handleFilterChange = () => {
    const { status } = filters.getFieldsValue();
    fetchInstances(1, pagination.pageSize, status);
  };

  const getStatusTag = (status: string) => {
    switch (status) {
      case 'running':
        return <Tag color="blue">运行中</Tag>;
      case 'stopped':
        return <Tag>已停止</Tag>;
      case 'expired':
        return <Tag color="orange">已过期</Tag>;
      default:
        return <Tag>{status}</Tag>;
    }
  };

  return (
    <div>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        实例管理
      </Typography.Title>

      <Form form={filters} layout="inline" style={{ margin: '16px 0' }} onValuesChange={handleFilterChange}>
        <Form.Item name="status">
          <Select placeholder="状态" allowClear style={{ width: 160 }} options={statusOptions} />
        </Form.Item>
        <Form.Item>
          <Button onClick={() => { filters.resetFields(); fetchInstances(); }}>
            重置
          </Button>
        </Form.Item>
        <Form.Item>
          <Button type="primary" icon={<ReloadOutlined />} onClick={() => fetchInstances(pagination.current, pagination.pageSize)}>
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
        expandable={{
          expandedRowKeys,
          onExpandedRowsChange: (keys) => setExpandedRowKeys(keys as string[]),
          expandedRowRender: (record) => (
            <InstanceStatsPanel instanceId={record.id} status={record.status} />
          ),
          rowExpandable: (record) => record.status === 'running',
        }}
        columns={[
          {
            title: '实例ID',
            dataIndex: 'id',
            width: 180,
            render: (v: string) => <Typography.Text copyable={{ text: v }}>{v.substring(0, 8)}...</Typography.Text>
          },
          {
            title: '题目',
            dataIndex: 'challenge_title',
            render: (v: string) => v || '-'
          },
          {
            title: '用户',
            dataIndex: 'user_id',
            width: 140,
            render: (v: string) => v?.substring(0, 12) || '-'
          },
          {
            title: '端口',
            dataIndex: 'port',
            width: 80,
          },
          {
            title: '状态',
            dataIndex: 'status',
            width: 100,
            render: (v: string) => getStatusTag(v),
          },
          {
            title: '过期时间',
            dataIndex: 'expires_at',
            width: 180,
            render: (v: string) => v ? new Date(v).toLocaleString() : '-',
          },
          {
            title: '创建时间',
            dataIndex: 'created_at',
            width: 180,
            render: (v: string) => new Date(v).toLocaleString(),
          },
          {
            title: '操作',
            key: 'actions',
            width: 100,
            render: (_: unknown, record: Instance) => (
              <Button
                size="small"
                onClick={() => {
                  setSelected(record);
                  setDrawerOpen(true);
                }}
              >
                详情
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
        title="实例详情"
        width={720}
      >
        {!selected ? null : (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <div>
              <Typography.Text strong>实例ID：</Typography.Text> {selected.id}
            </div>
            <div>
              <Typography.Text strong>题目：</Typography.Text> {selected.challenge_title || selected.challenge_id}
            </div>
            <div>
              <Typography.Text strong>用户：</Typography.Text> {selected.user_id || '-'}
            </div>
            <div>
              <Typography.Text strong>状态：</Typography.Text> {getStatusTag(selected.status)}
            </div>
            <div>
              <Typography.Text strong>容器ID：</Typography.Text> {selected.container_id || '-'}
            </div>
            <div>
              <Typography.Text strong>Docker主机ID：</Typography.Text> {selected.docker_host_id || '-'}
            </div>
            <div>
              <Typography.Text strong>端口：</Typography.Text> {selected.port ?? '-'}
            </div>
            <div>
              <Typography.Text strong>过期时间：</Typography.Text> {selected.expires_at ? new Date(selected.expires_at).toLocaleString() : '-'}
            </div>

            {/* 资源监控 */}
            {selected.status === 'running' && (
              <div style={{ marginTop: 16 }}>
                <Typography.Title level={5}>资源监控</Typography.Title>
                <InstanceStatsPanel instanceId={selected.id} status={selected.status} />
              </div>
            )}

            {/* 容器日志 */}
            <InstanceLogsPanel instanceId={selected.id} containerId={selected.container_id} />
          </Space>
        )}
      </Drawer>
    </div>
  );
};

export default AdminInstances;

import React, { useEffect, useState } from 'react';
import { Card, Table, Tag, Form, Input, Select, DatePicker, Space, Button, Statistic, Row, Col, Typography, Drawer, Descriptions, Badge } from 'antd';
import { SearchOutlined, ReloadOutlined, FilterOutlined, BugOutlined, GlobalOutlined, ClockCircleOutlined, BarsOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import logsApi from '../../../api/logs';
import type { APILog, LogStats } from '../../../api/logs';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Text } = Typography;

const Logs: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [logs, setLogs] = useState<APILog[]>([]);
    const [total, setTotal] = useState(0);
    const [stats, setStats] = useState<LogStats | null>(null);
    const [queryParams, setQueryParams] = useState({
        page: 1,
        page_size: 20,
        path: '',
        method: undefined as string | undefined,
        status_min: undefined as number | undefined,
        status_max: undefined as number | undefined,
        start_time: undefined as string | undefined,
        end_time: undefined as string | undefined,
        trace_id: '',
    });

    const [detailVisible, setDetailVisible] = useState(false);
    const [currentLog, setCurrentLog] = useState<APILog | null>(null);

    // 加载统计数据
    const fetchStats = async () => {
        try {
            const data = await logsApi.getLogStats();
            setStats(data);
        } catch (error) {
            console.error('Failed to fetch stats:', error);
        }
    };

    // 加载日志列表
    const fetchLogs = async () => {
        setLoading(true);
        try {
            const data = await logsApi.getLogs(queryParams);
            setLogs(data.list || []);
            setTotal(data.total);
        } catch (error) {
            console.error('Failed to fetch logs:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStats();
        fetchLogs();
    }, [queryParams]);

    const handleSearch = (values: any) => {
        const { timeRange, statusType, ...rest } = values;

        let status_min, status_max;
        if (statusType === 'error') {
            status_min = 400;
            status_max = 599;
        } else if (statusType === 'success') {
            status_min = 200;
            status_max = 299;
        }

        setQueryParams({
            ...queryParams,
            ...rest,
            page: 1,
            status_min,
            status_max,
            start_time: timeRange ? timeRange[0].toISOString() : undefined,
            end_time: timeRange ? timeRange[1].toISOString() : undefined,
        });
    };

    const columns: ColumnsType<APILog> = [
        {
            title: '时间',
            dataIndex: 'created_at',
            key: 'created_at',
            width: 170,
            render: (text) => <Text type="secondary" style={{ fontSize: 13 }}>{dayjs(text).format('YYYY-MM-DD HH:mm:ss')}</Text>,
        },
        {
            title: 'Method',
            dataIndex: 'method',
            key: 'method',
            width: 90,
            align: 'center',
            render: (text) => {
                let color = 'default';
                if (text === 'GET') color = 'blue';
                if (text === 'POST') color = 'green';
                if (text === 'DELETE') color = 'red';
                if (text === 'PUT') color = 'orange';
                return <Tag color={color} style={{ width: 60, textAlign: 'center', marginRight: 0, fontWeight: 600 }}>{text}</Tag>;
            },
        },
        {
            title: 'Path',
            dataIndex: 'path',
            key: 'path',
            render: (text) => <Text code copyable={{ text }}>{text}</Text>,
        },
        {
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
            width: 100,
            align: 'center',
            render: (status) => {
                const isError = status >= 400;
                return <Badge status={isError ? 'error' : 'success'} text={status} />;
            },
        },
        {
            title: 'Latency',
            dataIndex: 'latency_ms',
            key: 'latency_ms',
            width: 100,
            align: 'right',
            render: (val) => {
                if (val > 2000) return <Text type="danger" strong>{val} ms</Text>;
                if (val > 500) return <Text type="warning" strong>{val} ms</Text>;
                return <Text type="success" strong>{val} ms</Text>;
            },
            sorter: (a, b) => a.latency_ms - b.latency_ms,
        },
        {
            title: 'IP',
            dataIndex: 'ip',
            key: 'ip',
            width: 130,
            render: (text) => <Text style={{ fontSize: 13 }}>{text}</Text>,
        },
        {
            title: '操作',
            key: 'action',
            width: 80,
            align: 'center',
            render: (_, record) => (
                <Button type="link" size="small" onClick={() => { setCurrentLog(record); setDetailVisible(true); }}>
                    详情
                </Button>
            ),
        },
    ];

    return (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            {/* 统计卡片 */}
            <Row gutter={16}>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic
                            title="总请求数 (Total)"
                            value={stats?.total_requests}
                            prefix={<GlobalOutlined />}
                            valueStyle={{ color: '#1677ff' }}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic
                            title="错误请求 (Errors)"
                            value={stats?.error_requests}
                            valueStyle={{ color: '#cf1322' }}
                            prefix={<BugOutlined />}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic
                            title="平均延迟 (Avg Latency)"
                            value={stats?.avg_latency_ms}
                            precision={2}
                            suffix="ms"
                            prefix={<ClockCircleOutlined />}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card bordered={false} hoverable>
                        <Statistic
                            title="今日请求 (Today)"
                            value={stats?.today_requests}
                            prefix={<BarsOutlined />}
                            suffix={<span style={{ fontSize: 14, color: '#999', marginLeft: 8 }}>
                                (Errors: <span style={{ color: '#cf1322' }}>{stats?.today_errors}</span>)
                            </span>}
                        />
                    </Card>
                </Col>
            </Row>

            {/* 筛选栏 & 表格 */}
            <Card bordered={false} bodyStyle={{ padding: '0 24px 24px 24px' }}>
                <div style={{ padding: '24px 0', display: 'flex', justifyContent: 'space-between' }}>
                    <Form layout="inline" onFinish={handleSearch} style={{ flex: 1 }}>
                        <Form.Item name="path">
                            <Input placeholder="搜索路径 / Path" allowClear prefix={<SearchOutlined />} style={{ width: 200 }} />
                        </Form.Item>
                        <Form.Item name="method">
                            <Select placeholder="Method" allowClear style={{ width: 100 }}>
                                <Select.Option value="GET">GET</Select.Option>
                                <Select.Option value="POST">POST</Select.Option>
                                <Select.Option value="PUT">PUT</Select.Option>
                                <Select.Option value="DELETE">DELETE</Select.Option>
                            </Select>
                        </Form.Item>
                        <Form.Item name="statusType">
                            <Select placeholder="状态" allowClear style={{ width: 100 }}>
                                <Select.Option value="success">成功 (2xx)</Select.Option>
                                <Select.Option value="error">失败 (Err)</Select.Option>
                            </Select>
                        </Form.Item>
                        <Form.Item name="timeRange">
                            <RangePicker showTime />
                        </Form.Item>
                        <Form.Item>
                            <Button type="primary" htmlType="submit" icon={<FilterOutlined />}>筛选</Button>
                        </Form.Item>
                    </Form>
                    <Space>
                        <Button icon={<ReloadOutlined />} onClick={() => { fetchLogs(); fetchStats(); }}>刷新</Button>
                    </Space>
                </div>

                <Table
                    columns={columns}
                    dataSource={logs}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                        current: queryParams.page,
                        pageSize: queryParams.page_size,
                        total: total,
                        showSizeChanger: true,
                        showTotal: (total) => `共 ${total} 条记录`,
                        onChange: (page, pageSize) => setQueryParams({ ...queryParams, page, page_size: pageSize }),
                    }}
                />
            </Card>

            {/* 详情抽屉 */}
            <Drawer
                title="API 请求详情"
                width={700}
                onClose={() => setDetailVisible(false)}
                open={detailVisible}
            >
                {currentLog && (
                    <Space direction="vertical" size="large" style={{ width: '100%' }}>
                        <Descriptions bordered column={1} size="small" labelStyle={{ width: 120 }}>
                            <Descriptions.Item label="Trace ID">
                                <Text copyable>{currentLog.trace_id}</Text>
                            </Descriptions.Item>
                            <Descriptions.Item label="请求路径">
                                <Space>
                                    <Tag color="brand">{currentLog.method}</Tag>
                                    <Text code>{currentLog.path}</Text>
                                </Space>
                            </Descriptions.Item>
                            <Descriptions.Item label="状态码">
                                <Badge
                                    status={currentLog.status >= 400 ? 'error' : 'success'}
                                    text={<Text strong>{currentLog.status}</Text>}
                                />
                            </Descriptions.Item>
                            <Descriptions.Item label="响应耗时">
                                {currentLog.latency_ms} ms
                            </Descriptions.Item>
                            <Descriptions.Item label="请求时间">
                                {dayjs(currentLog.created_at).format('YYYY-MM-DD HH:mm:ss.SSS')}
                            </Descriptions.Item>
                            <Descriptions.Item label="客户端 IP">
                                {currentLog.ip}
                            </Descriptions.Item>
                            {currentLog.user_id && (
                                <Descriptions.Item label="User ID">
                                    <Text copyable>{currentLog.user_id}</Text>
                                </Descriptions.Item>
                            )}
                            <Descriptions.Item label="User Agent">
                                <Text type="secondary" style={{ fontSize: 12 }}>{currentLog.user_agent}</Text>
                            </Descriptions.Item>
                        </Descriptions>

                        {/* Request Body */}
                        {currentLog.request_body && (
                            <Card title="请求体 / Request Body" size="small" type="inner" bodyStyle={{ padding: 0 }}>
                                <div style={{
                                    padding: 12,
                                    background: '#fafafa',
                                    border: '1px solid #f0f0f0',
                                    fontFamily: 'Menlo, Monaco, monospace',
                                    fontSize: 12,
                                    whiteSpace: 'pre-wrap',
                                    maxHeight: 300,
                                    overflow: 'auto',
                                    color: '#595959'
                                }}>
                                    {tryFormatJSON(currentLog.request_body)}
                                </div>
                            </Card>
                        )}

                        {/* Response Body */}
                        {currentLog.response_body && (
                            <Card title="响应体 / Response Body" size="small" type="inner" bodyStyle={{ padding: 0 }}>
                                <div style={{
                                    padding: 12,
                                    background: '#fafafa',
                                    border: '1px solid #f0f0f0',
                                    fontFamily: 'Menlo, Monaco, monospace',
                                    fontSize: 12,
                                    whiteSpace: 'pre-wrap',
                                    maxHeight: 400,
                                    overflow: 'auto',
                                    color: '#595959'
                                }}>
                                    {tryFormatJSON(currentLog.response_body)}
                                </div>
                            </Card>
                        )}

                        {/* Error Message (Stack Trace) */}
                        {currentLog.error_message && (
                            <Card title="错误堆栈 / Stack Trace" size="small" type="inner" bodyStyle={{ padding: 0 }}>
                                <div style={{
                                    padding: 12,
                                    background: '#1e1e1e',
                                    color: '#ff7875',
                                    fontFamily: 'Menlo, Monaco, monospace',
                                    fontSize: 12,
                                    whiteSpace: 'pre-wrap',
                                    maxHeight: 400,
                                    overflow: 'auto'
                                }}>
                                    {currentLog.error_message}
                                </div>
                            </Card>
                        )}
                    </Space>
                )}
            </Drawer>
        </Space>
    );
};

// Helper: 尝试格式化 JSON
const tryFormatJSON = (str: string) => {
    try {
        const obj = JSON.parse(str);
        return JSON.stringify(obj, null, 2);
    } catch (e) {
        return str;
    }
};

export default Logs;

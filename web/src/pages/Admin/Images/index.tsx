import React, { useEffect, useState } from 'react';
import { Button, Table, Typography, message, Upload, Space, Alert, Progress, Card, Tag, Modal, Form, Input, Popconfirm } from 'antd';
import { UploadOutlined, ReloadOutlined, CloudSyncOutlined } from '@ant-design/icons';

import { getImages, uploadImage, syncImagesFromRegistry, deleteImage, type DockerImage } from '../../../api/admin';

const AdminImages: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [syncLoading, setSyncLoading] = useState(false);
    const [images, setImages] = useState<DockerImage[]>([]);
    const [error, setError] = useState<string | null>(null);
    const [uploadModalVisible, setUploadModalVisible] = useState(false);
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [form] = Form.useForm();

    const fetchImages = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await getImages();
            setImages(data);
        } catch (err: any) {
            setError(err.message || '获取镜像列表失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchImages();
    }, []);

    const beforeUpload = (file: File) => {
        const isTarOrGz = file.name.endsWith('.tar') || file.name.endsWith('.tar.gz');
        if (!isTarOrGz) {
            message.error('仅支持 .tar 或 .tar.gz 格式');
            return Upload.LIST_IGNORE;
        }
        setSelectedFile(file);
        setUploadModalVisible(true);
        form.resetFields();
        return false; // 阻止自动上传
    };

    const handleModalOk = async () => {
        if (!selectedFile) return;

        try {
            const values = await form.validateFields();
            setUploadModalVisible(false);
            setUploading(true);
            setUploadProgress(0);

            const result = await uploadImage(selectedFile, values.tag, (percent) => {
                setUploadProgress(percent);
            });

            message.success(`镜像导入成功: ${result.image_name}`);
            fetchImages(); // 刷新列表
        } catch (err: any) {
            const errMsg = err.message || '上传失败';
            message.error(errMsg);
        } finally {
            setUploading(false);
            setUploadProgress(0);
            setSelectedFile(null);
        }
    };

    const handleModalCancel = () => {
        setUploadModalVisible(false);
        setSelectedFile(null);
    };

    const handleSync = async () => {
        setSyncLoading(true);
        try {
            const result = await syncImagesFromRegistry();
            message.success(`同步完成，新增 ${result.synced_count} 个镜像`);
            fetchImages();
        } catch (err: any) {
            message.error(err.message || '同步失败');
        } finally {
            setSyncLoading(false);
        }
    };

    const handleDelete = async (id: string) => {
        try {
            await deleteImage(id);
            message.success('删除成功');
            fetchImages();
        } catch (err: any) {
            message.error(err.message || '删除失败');
        }
    };

    return (
        <div>
            <Typography.Title level={3} style={{ marginTop: 0 }}>
                镜像管理
            </Typography.Title>

            <Card style={{ marginBottom: 16 }}>
                <Space size="middle" wrap>
                    <Upload
                        accept=".tar,.gz"
                        showUploadList={false}
                        customRequest={() => { }} // 由于 beforeUpload 返回 false，这里不会被调用
                        beforeUpload={beforeUpload}
                        disabled={uploading}
                    >
                        <Button
                            type="primary"
                            icon={<UploadOutlined />}
                            loading={uploading}
                        >
                            {uploading ? '导入中...' : '上传镜像'}
                        </Button>
                    </Upload>

                    {uploading && (
                        <Progress percent={uploadProgress} size="small" style={{ width: 150 }} />
                    )}

                    <Button
                        icon={<CloudSyncOutlined />}
                        onClick={handleSync}
                        loading={syncLoading}
                    >
                        从 Registry 同步
                    </Button>

                    <Button icon={<ReloadOutlined />} onClick={fetchImages}>
                        刷新
                    </Button>
                </Space>

                <Typography.Paragraph type="secondary" style={{ marginTop: 12, marginBottom: 0 }}>
                    支持 .tar 或 .tar.gz 格式的 Docker 镜像包，上传后自动导入到本地 Registry。
                </Typography.Paragraph>
            </Card>

            {error && <Alert message={error} type="error" style={{ marginBottom: 16 }} />}

            <Table
                rowKey="id"
                dataSource={images}
                loading={loading}
                pagination={{ pageSize: 20 }}
                columns={[
                    {
                        title: '镜像名',
                        dataIndex: 'name',
                        render: (name: string, record: DockerImage) => (
                            <Typography.Text copyable={{ text: `${record.registry}/${name}:${record.tag}` }}>
                                {name}
                            </Typography.Text>
                        ),
                    },
                    {
                        title: '标签',
                        dataIndex: 'tag',
                        width: 120,
                        render: (tag: string) => <Tag color="blue">{tag}</Tag>,
                    },
                    {
                        title: 'Registry',
                        dataIndex: 'registry',
                        width: 160,
                    },
                    {
                        title: '描述',
                        dataIndex: 'description',
                        ellipsis: true,
                    },
                    {
                        title: '状态',
                        dataIndex: 'is_available',
                        width: 100,
                        render: (avail: boolean) =>
                            avail ? <Tag color="green">可用</Tag> : <Tag color="red">不可用</Tag>,
                    },
                    {
                        title: '创建时间',
                        dataIndex: 'created_at',
                        width: 180,
                        render: (v: string) => v ? new Date(v).toLocaleString() : '-',
                    },
                    {
                        title: '操作',
                        key: 'action',
                        width: 100,
                        render: (_: any, record: DockerImage) => (
                            <Popconfirm
                                title="确定删除此镜像吗？"
                                description="这将从系统和 Registry 中移除该镜像，请谨慎操作。"
                                onConfirm={() => handleDelete(record.id)}
                                okText="确定"
                                cancelText="取消"
                            >
                                <Button type="link" danger size="small">
                                    删除
                                </Button>
                            </Popconfirm>
                        ),
                    },
                ]}
            />

            <Modal
                title="上传镜像"
                open={uploadModalVisible}
                onOk={handleModalOk}
                onCancel={handleModalCancel}
                okText="开始上传"
                cancelText="取消"
            >
                <Form form={form} layout="vertical">
                    <Form.Item label="文件名">
                        <Input value={selectedFile?.name} disabled />
                    </Form.Item>
                    <Form.Item
                        name="tag"
                        label="自定义标签（可选）"
                        tooltip="如果不填写，将使用镜像中原有的标签或 'latest'"
                    >
                        <Input placeholder="例如: v1.0, latest" />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default AdminImages;

import React, { useEffect, useState } from 'react';
import { Button, Table, Typography, message, Upload, Space, Alert, Progress, Card, Tag } from 'antd';
import { UploadOutlined, ReloadOutlined, CloudSyncOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import { getImages, uploadImage, syncImagesFromRegistry, type DockerImage } from '../../../api/admin';

const AdminImages: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [syncLoading, setSyncLoading] = useState(false);
    const [images, setImages] = useState<DockerImage[]>([]);
    const [error, setError] = useState<string | null>(null);

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

    const handleUpload: UploadProps['customRequest'] = async (options) => {
        const { file, onSuccess, onError } = options;
        setUploading(true);
        setUploadProgress(0);

        try {
            const result = await uploadImage(file as File, (percent) => {
                setUploadProgress(percent);
            });
            message.success(`镜像导入成功: ${result.image_name}`);
            onSuccess?.(result, undefined as any);
            fetchImages(); // 刷新列表
        } catch (err: any) {
            const errMsg = err.message || '上传失败';
            message.error(errMsg);
            onError?.(err);
        } finally {
            setUploading(false);
            setUploadProgress(0);
        }
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
                        customRequest={handleUpload}
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
                ]}
            />
        </div>
    );
};

export default AdminImages;

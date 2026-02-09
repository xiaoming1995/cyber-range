import React, { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, InputNumber, Select, Space, Switch, Tabs, Tooltip, Typography, message } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ChallengeCategory, ChallengeDifficulty, ChallengeStatus } from '../../../admin/types';
import { createChallenge, getImages, type DockerImage } from '../../../api/admin';
import { getAdminToken } from '../../../admin/auth';
import RichTextEditor from '../../../components/RichTextEditor';
import axios from 'axios';

const categoryOptions: { label: string; value: ChallengeCategory }[] = [
  { label: 'Web', value: 'Web' },
  { label: 'Pwn', value: 'Pwn' },
  { label: 'Crypto', value: 'Crypto' },
  { label: 'Reverse', value: 'Reverse' },
  { label: 'Misc', value: 'Misc' },
];

const difficultyOptions: { label: string; value: ChallengeDifficulty }[] = [
  { label: 'Easy', value: 'Easy' },
  { label: 'Medium', value: 'Medium' },
  { label: 'Hard', value: 'Hard' },
];

const statusOptions: { label: string; value: ChallengeStatus }[] = [
  { label: '上架', value: 'published' },
  { label: '下架', value: 'unpublished' },
];

interface DockerHost {
  id: string;
  name: string;
  host: string;
  enabled: boolean;
  is_default: boolean;
}

const AdminChallengeNew: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [saving, setSaving] = useState(false);
  const [dockerHosts, setDockerHosts] = useState<DockerHost[]>([]);
  const [loadingHosts, setLoadingHosts] = useState(false);
  const [dockerImages, setDockerImages] = useState<DockerImage[]>([]);
  const [loadingImages, setLoadingImages] = useState(false);

  // 获取 Docker 主机列表
  useEffect(() => {
    const fetchDockerHosts = async () => {
      try {
        setLoadingHosts(true);
        const token = getAdminToken();
        if (!token) return;

        const response = await axios.get('/api/admin/docker-hosts', {
          headers: { Authorization: `Bearer ${token}` },
        });

        if (response.data.code === 200) {
          const hosts = response.data.data;
          setDockerHosts(hosts);

          // 设置默认主机
          const defaultHost = hosts.find((h: DockerHost) => h.is_default);
          if (defaultHost) {
            form.setFieldsValue({ docker_host_id: defaultHost.id });
          }
        }
      } catch (error) {
        console.error('Failed to fetch docker hosts:', error);
      } finally {
        setLoadingHosts(false);
      }
    };

    fetchDockerHosts();
  }, [form]);

  // 获取镜像列表
  useEffect(() => {
    const fetchImages = async () => {
      try {
        setLoadingImages(true);
        const images = await getImages();
        setDockerImages(images);
      } catch (error) {
        console.error('Failed to fetch docker images:', error);
        message.error('加载镜像列表失败');
      } finally {
        setLoadingImages(false);
      }
    };

    fetchImages();
  }, []);

  // 镜像选择时自动填充推荐资源
  const handleImageChange = (imageId: string) => {
    const selectedImage = dockerImages.find(img => img.id === imageId);
    if (selectedImage) {
      // 自动填充推荐资源 (仅当当前值为空或0时)
      const currentMemory = form.getFieldValue('memory_limit');
      const currentCPU = form.getFieldValue('cpu_limit');
      if (!currentMemory && selectedImage.recommended_memory) {
        form.setFieldValue('memory_limit', Math.round(selectedImage.recommended_memory / 1024 / 1024)); // 转换为 MB
      }
      if (!currentCPU && selectedImage.recommended_cpu) {
        form.setFieldValue('cpu_limit', selectedImage.recommended_cpu);
      }
    }
  };

  const tabs = useMemo(
    () => [
      {
        key: 'basic',
        label: '基本信息',
        forceRender: true,
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
              <Input />
            </Form.Item>
            <Space style={{ width: '100%' }} size="large" wrap>
              <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]} style={{ minWidth: 220 }}>
                <Select options={categoryOptions} />
              </Form.Item>
              <Form.Item
                name="difficulty"
                label="难度"
                rules={[{ required: true, message: '请选择难度' }]}
                style={{ minWidth: 220 }}
              >
                <Select options={difficultyOptions} />
              </Form.Item>
              <Form.Item
                name="points"
                label="分值"
                rules={[{ required: true, message: '请输入分值' }]}
                style={{ minWidth: 220 }}
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Space>
            <Form.Item
              name="image_id"
              label="Docker 镜像"
              rules={[{ required: true, message: '请选择镜像' }]}
              tooltip="选择题目使用的 Docker 镜像"
            >
              <Select
                loading={loadingImages}
                placeholder="选择 Docker 镜像"
                showSearch
                filterOption={(input, option) =>
                  (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                }
                onChange={handleImageChange}
                options={dockerImages.map(img => ({
                  label: `${img.name}:${img.tag}${img.description ? ` - ${img.description}` : ''}`,
                  value: img.id,
                }))}
                notFoundContent={loadingImages ? '加载中...' : '暂无镜像，请先导入镜像'}
              />
            </Form.Item>
            <Form.Item
              name="docker_host_id"
              label="Docker 主机"
              rules={[{ required: true, message: '请选择 Docker 主机' }]}
              tooltip="选择此题目容器运行的 Docker 主机"
            >
              <Select
                loading={loadingHosts}
                placeholder="选择 Docker 主机"
                options={dockerHosts.map(host => ({
                  label: `${host.name} ${host.enabled ? '' : '(已禁用)'}${host.is_default ? ' [默认]' : ''}`,
                  value: host.id,
                  disabled: !host.enabled
                }))}
              />
            </Form.Item>
          </Space>
        ),
      },
      {
        key: 'content',
        label: '内容',
        forceRender: true,
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Form.Item name="descriptionHtml" label="题目描述" rules={[{ required: true, message: '请输入题目描述' }]} valuePropName="value">
              <RichTextEditor placeholder="输入题目描述…" minHeight={260} />
            </Form.Item>
            <Form.Item name="hintHtml" label="提示" valuePropName="value">
              <RichTextEditor placeholder="输入提示（可选）…" minHeight={160} />
            </Form.Item>
          </Space>
        ),
      },
      {
        key: 'runtime',
        label: '运行配置',
        forceRender: true,
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Space style={{ width: '100%' }} size="large" wrap>
              <Form.Item
                name="port"
                label="容器端口"
                rules={[{ required: true, message: '请输入端口' }]}
                style={{ minWidth: 220 }}
              >
                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item
                name="memory_limit"
                label="内存限制 (MB)"
                tooltip="容器最大可用内存，留空表示使用默认值(128MB)"
                style={{ minWidth: 220 }}
              >
                <InputNumber min={0} max={8192} placeholder="128" style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item
                name="cpu_limit"
                label="CPU 限制 (核)"
                tooltip="容器最大可用CPU核心数，留空表示使用默认值(0.5核)"
                style={{ minWidth: 220 }}
              >
                <InputNumber min={0} max={8} step={0.1} placeholder="0.5" style={{ width: '100%' }} />
              </Form.Item>
            </Space>
            <Form.Item
              name="privileged"
              label={
                <Space>
                  特权模式
                  <Tooltip title="启用后容器将以 --privileged 模式运行，拥有完整权限。仅在必要时启用。">
                    <ExclamationCircleOutlined style={{ color: '#faad14' }} />
                  </Tooltip>
                </Space>
              }
              valuePropName="checked"
            >
              <Switch checkedChildren="开启" unCheckedChildren="关闭" />
            </Form.Item>
            <Form.Item name="status" label="状态" rules={[{ required: true, message: '请选择状态' }]}>
              <Select options={statusOptions} style={{ maxWidth: 220 }} />
            </Form.Item>
          </Space>
        ),
      },
      {
        key: 'flag',
        label: 'Flag',
        forceRender: true,
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Form.Item name="flag" label="Flag" rules={[{ required: true, message: '请输入 Flag' }]}>
              <Input.Password visibilityToggle={false} placeholder="flag{...}" />
            </Form.Item>
          </Space>
        ),
      },
    ],
    [dockerImages, loadingImages, dockerHosts, loadingHosts, form, handleImageChange],
  );

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography.Title level={3} style={{ marginTop: 0 }}>
          新建挑战
        </Typography.Title>
        <Space>
          <Button onClick={() => navigate('/admin/challenges')}>取消</Button>
          <Button
            loading={saving}
            onClick={async () => {
              try {
                const values = await form.validateFields();
                setSaving(true);
                const token = getAdminToken();
                if (!token) {
                  message.error('请先登录');
                  navigate('/admin/login');
                  return;
                }
                // 内存单位转换: MB -> Bytes
                const submitData = { ...values };
                if (submitData.memory_limit) {
                  submitData.memory_limit = Math.round(submitData.memory_limit * 1024 * 1024);
                }
                await createChallenge(submitData);
                message.success('已创建');
                navigate('/admin/challenges');
              } catch (error: any) {
                message.error(error.message || '创建失败');
              } finally {
                setSaving(false);
              }
            }}
          >
            保存
          </Button>
          <Button
            type="primary"
            loading={saving}
            onClick={async () => {
              try {
                const values = await form.validateFields();
                setSaving(true);
                const token = getAdminToken();
                if (!token) {
                  message.error('请先登录');
                  navigate('/admin/login');
                  return;
                }
                // 内存单位转换: MB -> Bytes
                const submitData = { ...values, status: 'published' };
                if (submitData.memory_limit) {
                  submitData.memory_limit = Math.round(submitData.memory_limit * 1024 * 1024);
                }
                await createChallenge(submitData);
                message.success('已创建并上架');
                navigate('/admin/challenges');
              } catch (error: any) {
                message.error(error.message || '创建失败');
              } finally {
                setSaving(false);
              }
            }}
          >
            保存并上架
          </Button>
        </Space>
      </div>

      <Form
        layout="vertical"
        form={form}
        initialValues={{
          category: 'Web',
          difficulty: 'Easy',
          points: 100,
          image: '',
          docker_host_id: '', // 会在 useEffect 中设置默认值
          port: 80,
          memory_limit: undefined,
          cpu_limit: undefined,
          privileged: false,
          status: 'unpublished',
          descriptionHtml: '',
          hintHtml: '',
          flag: '',
        }}
      >
        <Tabs items={tabs} />
      </Form>
    </div>
  );
};

export default AdminChallengeNew;

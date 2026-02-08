import React, { useEffect, useMemo, useState } from 'react';
import { Button, Drawer, Form, Input, InputNumber, Modal, Select, Space, Switch, Tabs, Tooltip, Typography, message } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';

import type { AdminChallenge, ChallengeCategory, ChallengeDifficulty, ChallengeStatus } from '../../../admin/types';
import RichTextEditor from '../../../components/RichTextEditor';
import adminApi, { getImages, type DockerImage } from '../../../api/admin';

type Props = {
  open: boolean;
  challenge: AdminChallenge | null;
  onClose: () => void;
  onSave: (id: string, patch: Partial<AdminChallenge>) => Promise<void>;
};

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

const ChallengeEditDrawer: React.FC<Props> = ({ open, challenge, onClose, onSave }) => {
  const [form] = Form.useForm();
  const [revealFlag, setRevealFlag] = useState(false);
  const [dockerHosts, setDockerHosts] = useState<DockerHost[]>([]);
  const [loadingHosts, setLoadingHosts] = useState(false);
  const [dockerImages, setDockerImages] = useState<DockerImage[]>([]);
  const [loadingImages, setLoadingImages] = useState(false);

  // 获取 Docker 主机列表
  useEffect(() => {
    if (!open) return;

    const fetchDockerHosts = async () => {
      try {
        setLoadingHosts(true);
        console.warn('[Debug] Fetching docker hosts via adminApi...');

        // 使用 adminApi，自动处理 baseURL 和 Token
        const response = await adminApi.get('/docker-hosts');

        console.warn('[Debug] API Response:', response.data);

        if (response.data.code === 200) {
          setDockerHosts(response.data.data);
          console.warn(`[Debug] Loaded ${response.data.data.length} hosts`);
        } else {
          message.error('加载 Docker 主机失败: ' + response.data.msg);
        }
      } catch (error) {
        console.error('[Debug] Failed to fetch docker hosts:', error);
        // 如果是 401，adminApi 拦截器会自动跳转登录，这里不需要额外处理
        if ((error as any)?.response?.status !== 401) {
          message.error('加载 Docker 主机失败');
        }
      } finally {
        setLoadingHosts(false);
      }
    };

    fetchDockerHosts();
  }, [open]);

  // 获取镜像列表
  useEffect(() => {
    if (!open) return;

    const fetchImages = async () => {
      try {
        setLoadingImages(true);
        const images = await getImages();
        setDockerImages(images);
      } catch (error) {
        console.error('Failed to fetch docker images:', error);
      } finally {
        setLoadingImages(false);
      }
    };

    fetchImages();
  }, [open]);

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

  useEffect(() => {
    if (!open || !challenge) return;

    // 计算镜像 ID：优先使用 image_id，如果没有则尝试通过 image 名称匹配
    let resolvedImageId = (challenge as any).image_id;
    if (!resolvedImageId && challenge.image && dockerImages.length > 0) {
      const targetImage = challenge.image.trim();
      const found = dockerImages.find(img => {
        const fullNameTag = `${img.name}:${img.tag}`; // ntc:latest
        const fullPath = `${img.registry}/${img.name}:${img.tag}`; // localhost:5000/ntc:latest
        const fullPathNoTag = `${img.registry}/${img.name}`; // localhost:5000/ntc

        return img.name === targetImage ||
          fullNameTag === targetImage ||
          fullPath === targetImage ||
          fullPathNoTag === targetImage;
      });

      if (found) {
        resolvedImageId = found.id;
        console.log('[Debug] Auto-matched image:', targetImage, '->', found.id);
      } else {
        console.warn('[Debug] Failed to match image:', targetImage, 'Available:', dockerImages.map(i => i.name));
      }
    }

    // 计算 Docker Host ID: 如果题目没有指定 Host ID，尝试使用默认主机
    let resolvedHostId = (challenge as any).docker_host_id;
    if (!resolvedHostId && dockerHosts.length > 0) {
      const defaultHost = dockerHosts.find(h => h.is_default);
      if (defaultHost) {
        resolvedHostId = defaultHost.id;
      }
    }

    form.setFieldsValue({
      title: challenge.title,
      category: challenge.category,
      difficulty: challenge.difficulty,
      points: challenge.points,
      image: challenge.image,
      image_id: resolvedImageId,
      docker_host_id: resolvedHostId, // 使用计算出的 Host ID
      port: challenge.port,
      memory_limit: challenge.memory_limit ? Math.round(challenge.memory_limit / 1024 / 1024) : undefined,
      cpu_limit: challenge.cpu_limit || undefined,
      status: challenge.status,
      descriptionHtml: challenge.descriptionHtml,
      hintHtml: challenge.hintHtml,
      flag: challenge.flag,
      privileged: (challenge as any).privileged || false,
    });

    // 异步获取完整详情以确保数据最新 (特别是 privileged 字段)
    const fetchDetail = async () => {
      try {
        // 导入 getChallenge 防止循环依赖，或者直接用 adminApi
        // 由于上面 import 了 adminApi, 我们直接用 adminApi.get
        const res = await adminApi.get(`/challenges/${challenge.id}`);
        if (res.data.code === 200) {
          const detail = res.data.data;
          console.log('[Debug] Fetched full detail:', detail);
          // 更新 privileged 字段
          // 注意：如果详情接口返回的字段名是 privileged，则直接使用
          if (detail.privileged !== undefined) {
            form.setFieldsValue({
              privileged: detail.privileged
            });
          }
        }
      } catch (err) {
        console.error('[Debug] Failed to fetch challenge detail:', err);
      }
    };
    fetchDetail();

  }, [open, challenge, form, dockerImages, dockerHosts]);

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
                optionFilterProp="label"
                onChange={handleImageChange}
                options={dockerImages.map(img => ({
                  label: `${img.name}:${img.tag} ${img.description ? `- ${img.description}` : ''}`,
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
            <Form.Item
              name="descriptionHtml"
              label="题目描述"
              rules={[{ required: true, message: '请输入题目描述' }]}
              valuePropName="value"
            >
              <RichTextEditor placeholder="输入题目描述…" minHeight={240} />
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
          </Space>
        ),
      },
      {
        key: 'flag',
        label: 'Flag',
        forceRender: true,
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <div>
              <Typography.Text type="secondary">Flag 默认隐藏，查看需要二次确认。</Typography.Text>
            </div>
            <Space wrap>
              <Button
                onClick={() => {
                  Modal.confirm({
                    title: '查看 Flag',
                    content: '请确认你具备查看权限，避免泄露。',
                    okText: '确认查看',
                    cancelText: '取消',
                    onOk: () => setRevealFlag(true),
                  });
                }}
              >
                查看 Flag
              </Button>
              <Button
                disabled={!revealFlag}
                onClick={() => {
                  const text = form.getFieldValue('flag');
                  if (!text) return;
                  navigator.clipboard.writeText(text);
                  message.success('已复制');
                }}
              >
                复制
              </Button>
            </Space>
            <Form.Item name="flag" label="Flag" rules={[{ required: true, message: '请输入 Flag' }]}>
              {revealFlag ? <Input placeholder="flag{...}" /> : <Input.Password visibilityToggle={false} placeholder="flag{...}" />}
            </Form.Item>
          </Space>
        ),
      },
      {
        key: 'status',
        label: '状态',
        children: (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Form.Item name="status" label="上架状态">
              <Select options={statusOptions} />
            </Form.Item>
            <div>
              <Typography.Text type="secondary">上架后用户侧可见；下架后仅后台可见。</Typography.Text>
            </div>
          </Space>
        ),
      },
    ],
    [form, revealFlag, dockerHosts, loadingHosts],
  );

  if (!challenge) return null;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={`编辑挑战：${challenge.title}`}
      width={860}
      destroyOnClose
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Space>
            <Button onClick={onClose}>取消</Button>
            <Button
              type="primary"
              onClick={async () => {
                const values = await form.validateFields();
                // 转换数据格式
                const submitData = { ...values };

                // 内存单位转换: MB -> Bytes
                if (submitData.memory_limit) {
                  submitData.memory_limit = Math.round(submitData.memory_limit * 1024 * 1024);
                }

                // 确保 image_id 存在
                if (!submitData.image_id && submitData.image) {
                  // 如果只有 image 没有 image_id，这里可能需要反向查找？
                  // 但这里是保存，通常不需处理，或者在 form 中已经处理了。
                }

                // 如果 Form 中 docker_host_id 为 undefined，可能需要保留 undefined

                try {
                  await onSave(challenge.id, submitData);
                  message.success('已保存');
                  onClose();
                } catch (error) {
                  console.error('Save failed:', error);
                }
              }}
            >
              保存
            </Button>
          </Space>
        </div>
      }
    >
      <Form layout="vertical" form={form}>
        <Tabs items={tabs} />
      </Form>
    </Drawer>
  );
};

export default ChallengeEditDrawer;

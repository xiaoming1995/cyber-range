import React, { useEffect, useState } from 'react';
import { 
  List, 
  Card, 
  Tag, 
  Button, 
  Space, 
  Typography, 
  Input, 
  message, 
  Modal, 
  Descriptions,
  Badge,
  Alert
} from 'antd';
import { 
  PlayCircleOutlined, 
  StopOutlined, 
  TrophyOutlined, 
  CheckCircleOutlined,
  CloseCircleOutlined
} from '@ant-design/icons';
import { 
  getChallenges, 
  startInstance, 
  stopInstance, 
  submitFlag 
} from '../../api/challenges';
import type { Challenge, Instance } from '../../api/challenges';

const { Title, Paragraph, Text } = Typography;
const { Search } = Input;

const Challenges: React.FC = () => {
  const [challenges, setChallenges] = useState<Challenge[]>([]);
  const [instances, setInstances] = useState<Record<string, Instance>>({});
  const [loading, setLoading] = useState<Record<string, boolean>>({});
  const [modalVisible, setModalVisible] = useState(false);
  const [activeChallenge, setActiveChallenge] = useState<Challenge | null>(null);
  
  // Flag submission state
  const [submitting, setSubmitting] = useState(false);
  const [lastResult, setLastResult] = useState<{correct: boolean, message: string} | null>(null);

  useEffect(() => {
    loadChallenges();
  }, []);

  const loadChallenges = async () => {
    try {
      const data = await getChallenges();
      setChallenges(data);
    } catch (error) {
      message.success('加载挑战失败');
      console.error(error);
    }
  };

  const handleStart = async (challenge: Challenge) => {
    setLoading(prev => ({ ...prev, [challenge.id]: true }));
    try {
      const instance = await startInstance(challenge.id);
      setInstances(prev => ({ ...prev, [challenge.id]: instance }));
      message.success(`已启动环境: ${challenge.title}`);
    } catch (error) {
      message.error('启动环境失败，请检查 Docker 是否运行');
      console.error(error);
    } finally {
      setLoading(prev => ({ ...prev, [challenge.id]: false }));
    }
  };

  const handleStop = async (challenge: Challenge) => {
    setLoading(prev => ({ ...prev, [challenge.id]: true }));
    try {
      await stopInstance(challenge.id);
      setInstances(prev => {
        const next = { ...prev };
        delete next[challenge.id];
        return next;
      });
      message.success('环境已停止');
    } catch (error) {
      message.error('停止环境失败');
      console.error(error);
    } finally {
      setLoading(prev => ({ ...prev, [challenge.id]: false }));
    }
  };

  const openDetail = (challenge: Challenge) => {
    setActiveChallenge(challenge);
    setModalVisible(true);
    setLastResult(null);
  };

  const handleFlagSubmit = async (flag: string) => {
    if (!activeChallenge) return;
    
    setSubmitting(true);
    try {
      const result = await submitFlag(activeChallenge.id, flag);
      setLastResult(result);
      if (result.correct) {
        message.success('恭喜！Flag 正确！');
      } else {
        message.error('Flag 错误，请重试。');
      }
    } catch (error) {
      message.error('提交失败');
      console.error(error);
    } finally {
      setSubmitting(false);
    }
  };

  const getDifficultyColor = (diff: string) => {
    switch (diff.toLowerCase()) {
      case 'easy': return 'green';
      case 'medium': return 'orange';
      case 'hard': return 'red';
      default: return 'blue';
    }
  };

  return (
    <div style={{ padding: '0 24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>网络安全实战训练</Title>
        <Paragraph>
          选择一个挑战开始你的学习之旅。启动环境，寻找漏洞，夺取 Flag。
        </Paragraph>
      </div>

      <List
        grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 3, xl: 4, xxl: 4 }}
        dataSource={challenges}
        renderItem={(item) => (
          <List.Item>
            <Card 
              hoverable
              actions={[
                <Button type="link" onClick={() => openDetail(item)}>查看详情</Button>,
                instances[item.id] ? (
                  <Button 
                    danger 
                    size="small" 
                    icon={<StopOutlined />} 
                    loading={loading[item.id]}
                    onClick={() => handleStop(item)}
                  >
                    停止
                  </Button>
                ) : (
                  <Button 
                    type="primary" 
                    size="small" 
                    icon={<PlayCircleOutlined />} 
                    loading={loading[item.id]}
                    onClick={() => handleStart(item)}
                  >
                    启动
                  </Button>
                )
              ]}
            >
              <Card.Meta
                avatar={<TrophyOutlined style={{ fontSize: 24, color: '#1890ff' }} />}
                title={item.title}
                description={
                  <Space direction="vertical" size={4}>
                    <Tag color={getDifficultyColor(item.difficulty)}>{item.difficulty}</Tag>
                    <Text type="secondary">{item.points} 分</Text>
                    {instances[item.id] && (
                      <Badge status="processing" text="运行中" />
                    )}
                  </Space>
                }
              />
            </Card>
          </List.Item>
        )}
      />

      <Modal
        title={activeChallenge?.title}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={800}
      >
        {activeChallenge && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            <Descriptions bordered column={1}>
              <Descriptions.Item label="分类">{activeChallenge.category}</Descriptions.Item>
              <Descriptions.Item label="难度">
                <Tag color={getDifficultyColor(activeChallenge.difficulty)}>{activeChallenge.difficulty}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="描述">
                {activeChallenge.description}
              </Descriptions.Item>
            </Descriptions>

            <Card title="环境控制" size="small">
              {instances[activeChallenge.id] ? (
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Alert
                    message="实例运行中"
                    description={
                      <span>
                        访问目标地址: <a href={`http://localhost:${instances[activeChallenge.id].port}`} target="_blank" rel="noreferrer">
                          http://localhost:{instances[activeChallenge.id].port}
                        </a>
                      </span>
                    }
                    type="success"
                    showIcon
                  />
                  <Button danger onClick={() => handleStop(activeChallenge)}>停止实例</Button>
                </Space>
              ) : (
                <div style={{ textAlign: 'center', padding: 16 }}>
                  <Button 
                    type="primary" 
                    size="large" 
                    icon={<PlayCircleOutlined />} 
                    onClick={() => handleStart(activeChallenge)}
                    loading={loading[activeChallenge.id]}
                  >
                    启动环境
                  </Button>
                </div>
              )}
            </Card>

            <Card title="提交 Flag" size="small">
              <Search
                placeholder="在此输入 Flag (例如 flag{...})"
                enterButton="提交"
                size="large"
                onSearch={handleFlagSubmit}
                loading={submitting}
                disabled={lastResult?.correct}
              />
              {lastResult && (
                <div style={{ marginTop: 16 }}>
                  {lastResult.correct ? (
                    <Alert message={lastResult.message} type="success" showIcon icon={<CheckCircleOutlined />} />
                  ) : (
                    <Alert message={lastResult.message} type="error" showIcon icon={<CloseCircleOutlined />} />
                  )}
                </div>
              )}
            </Card>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Challenges;

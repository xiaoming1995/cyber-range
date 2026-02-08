export type ChallengeDifficulty = 'Easy' | 'Medium' | 'Hard';

export type ChallengeStatus = 'published' | 'unpublished';

export type ChallengeCategory = 'Web' | 'Pwn' | 'Crypto' | 'Reverse' | 'Misc';

export interface AdminChallenge {
  id: string;
  title: string;
  category: ChallengeCategory;
  difficulty: ChallengeDifficulty;
  points: number;
  image: string;
  image_id?: string;      // 关联镜像 ID
  docker_host_id?: string;  // Docker 主机 ID
  port: number;
  memory_limit?: number;  // 内存限制(字节)
  cpu_limit?: number;     // CPU限制(核心数)
  privileged?: boolean;   // 特权模式
  status: ChallengeStatus;
  descriptionHtml: string;
  hintHtml?: string;
  flag: string;
  updatedAt: string;
  createdAt: string;
}

export type InstanceStatus = 'running' | 'stopped';

export interface AdminInstance {
  id: string;
  challengeId: string;
  challengeTitle: string;
  userDisplayName?: string;
  status: InstanceStatus;
  containerId?: string;
  image?: string;
  port?: number;
  createdAt: string;
  stoppedAt?: string;
}

export type SubmissionResult = 'correct' | 'wrong';

export interface AdminSubmission {
  id: string;
  challengeId: string;
  challengeTitle: string;
  userDisplayName: string;
  result: SubmissionResult;
  ip?: string;
  createdAt: string;
}


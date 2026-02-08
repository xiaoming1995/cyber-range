import { getJsonFromStorage, setJsonToStorage } from './storage';
import type { AdminChallenge, AdminInstance, AdminSubmission, ChallengeCategory, ChallengeDifficulty, ChallengeStatus } from './types';

const STORAGE_KEYS = {
  challenges: 'cyber_range_admin_challenges_v1',
  instances: 'cyber_range_admin_instances_v1',
  submissions: 'cyber_range_admin_submissions_v1',
} as const;

function nowIso(): string {
  return new Date().toISOString();
}

function newId(prefix: string): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return `${prefix}_${crypto.randomUUID()}`;
  }
  return `${prefix}_${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

function seedChallenges(): AdminChallenge[] {
  const now = nowIso();
  return [
    {
      id: '1',
      title: 'Nginx 基础挑战',
      category: 'Web',
      difficulty: 'Easy',
      points: 100,
      image: 'nginx:alpine',
      port: 80,
      status: 'published',
      descriptionHtml:
        '<h3>题目说明</h3><p>一个简单的 Web 服务挑战。请在环境变量或默认页面中找到 Flag。</p>',
      hintHtml: '<p>先访问首页，再看看是否存在环境变量或隐藏路径。</p>',
      flag: 'flag{welcome_to_cyber_range}',
      createdAt: now,
      updatedAt: now,
    },
    {
      id: '2',
      title: '神秘的 Python',
      category: 'Web',
      difficulty: 'Medium',
      points: 200,
      image: 'python:3.9-slim',
      port: 80,
      status: 'unpublished',
      descriptionHtml: '<h3>题目说明</h3><p>一个包含隐藏秘密的 Python Web 应用。</p>',
      hintHtml: '<p>注意路由枚举与常见备份文件。</p>',
      flag: 'flag{python_snake_charmer}',
      createdAt: now,
      updatedAt: now,
    },
  ];
}

function seedInstances(): AdminInstance[] {
  return [];
}

function seedSubmissions(): AdminSubmission[] {
  const base = Date.now();
  const iso = (t: number) => new Date(t).toISOString();
  return [
    {
      id: newId('sub'),
      challengeId: '1',
      challengeTitle: 'Nginx 基础挑战',
      userDisplayName: 'alice',
      result: 'wrong',
      ip: '127.0.0.1',
      createdAt: iso(base - 1000 * 60 * 3),
    },
    {
      id: newId('sub'),
      challengeId: '1',
      challengeTitle: 'Nginx 基础挑战',
      userDisplayName: 'alice',
      result: 'correct',
      ip: '127.0.0.1',
      createdAt: iso(base - 1000 * 60 * 2),
    },
    {
      id: newId('sub'),
      challengeId: '2',
      challengeTitle: '神秘的 Python',
      userDisplayName: 'bob',
      result: 'wrong',
      ip: '127.0.0.1',
      createdAt: iso(base - 1000 * 60),
    },
  ];
}

export function ensureAdminSeeded(): void {
  const challenges = getJsonFromStorage<AdminChallenge[] | null>(STORAGE_KEYS.challenges, null);
  if (!challenges || challenges.length === 0) {
    setJsonToStorage(STORAGE_KEYS.challenges, seedChallenges());
  }

  const instances = getJsonFromStorage<AdminInstance[] | null>(STORAGE_KEYS.instances, null);
  if (!instances) {
    setJsonToStorage(STORAGE_KEYS.instances, seedInstances());
  }

  const submissions = getJsonFromStorage<AdminSubmission[] | null>(STORAGE_KEYS.submissions, null);
  if (!submissions) {
    setJsonToStorage(STORAGE_KEYS.submissions, seedSubmissions());
  }
}

export function listChallenges(): AdminChallenge[] {
  ensureAdminSeeded();
  return getJsonFromStorage<AdminChallenge[]>(STORAGE_KEYS.challenges, []);
}

export function saveChallenges(items: AdminChallenge[]): void {
  setJsonToStorage(STORAGE_KEYS.challenges, items);
}

export function createChallenge(input: {
  title: string;
  category: ChallengeCategory;
  difficulty: ChallengeDifficulty;
  points: number;
  image: string;
  port: number;
  status: ChallengeStatus;
  descriptionHtml: string;
  hintHtml?: string;
  flag: string;
}): AdminChallenge {
  const now = nowIso();
  const challenge: AdminChallenge = {
    id: newId('chal'),
    createdAt: now,
    updatedAt: now,
    ...input,
  };
  const items = listChallenges();
  items.unshift(challenge);
  saveChallenges(items);
  return challenge;
}

export function updateChallenge(id: string, patch: Partial<Omit<AdminChallenge, 'id' | 'createdAt'>>): AdminChallenge | null {
  const items = listChallenges();
  const idx = items.findIndex((c) => c.id === id);
  if (idx < 0) return null;
  const updated: AdminChallenge = { ...items[idx], ...patch, updatedAt: nowIso() };
  items[idx] = updated;
  saveChallenges(items);
  return updated;
}

export function deleteChallenge(id: string): void {
  const items = listChallenges().filter((c) => c.id !== id);
  saveChallenges(items);
}

export function listInstances(): AdminInstance[] {
  ensureAdminSeeded();
  return getJsonFromStorage<AdminInstance[]>(STORAGE_KEYS.instances, []);
}

export function saveInstances(items: AdminInstance[]): void {
  setJsonToStorage(STORAGE_KEYS.instances, items);
}

export function listSubmissions(): AdminSubmission[] {
  ensureAdminSeeded();
  return getJsonFromStorage<AdminSubmission[]>(STORAGE_KEYS.submissions, []);
}

